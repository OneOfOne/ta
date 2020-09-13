package ta

import (
	"math"
	"sync"
)

// LockedStudy returns a thread-safe version of the study
func LockedStudy(s Study) Study {
	return &locked{Study: s}
}

// LockedMulti returns a thread-safe version of the multi variable study
func LockedMulti(s MultiVarStudy) MultiVarStudy {
	return &lockedMulti{MultiVarStudy: s}
}

type locked struct {
	Study
	m sync.Mutex
}

func (l *locked) Update(vs ...Decimal) Decimal {
	l.m.Lock()
	v := l.Study.Update(vs...)
	l.m.Unlock()
	return v
}

func (l *locked) ToMulti() (s MultiVarStudy, ok bool) {
	l.m.Lock()
	defer l.m.Unlock()
	if s, ok = l.Study.ToMulti(); ok {
		if _, isLocked := s.(*lockedMulti); !isLocked {
			s = &lockedMulti{MultiVarStudy: s}
		}
	}
	return
}

type lockedMulti struct {
	MultiVarStudy
	m sync.Mutex
}

func (l *lockedMulti) UpdateAll(vs ...Decimal) []Decimal {
	l.m.Lock()
	vs = l.MultiVarStudy.UpdateAll(vs...)
	l.m.Unlock()
	return vs
}

func (l *lockedMulti) ToStudy() (s Study, ok bool) {
	l.m.Lock()
	defer l.m.Unlock()
	if s, ok = l.MultiVarStudy.ToStudy(); ok {
		if _, isLocked := s.(*locked); !isLocked {
			s = &locked{Study: s}
		}
	}
	return
}

// Study represents a TA study that supports live updates
type Study interface {
	// Update depends on the study, must studies can accept multiple values and returns the results
	// however some studies, like VWAP, expects exactly 2 values [volume, price]
	Update(values ...Decimal) Decimal

	// Len returns the period the study was created with
	Len() int

	// ToMulti can be used to convert the study to a multi-variable study if it's supported
	ToMulti() (MultiVarStudy, bool)
}

// StudyWithSetup is a study that supports a setup function for the initial data set
type StudyWithSetup interface {
	Study
	// Setup takes in multiple TAs and returns the result
	// len(res) == len(out)
	Setup(tas ...*TA) (out []*TA)
}

type noMulti struct{}

func (noMulti) ToMulti() (MultiVarStudy, bool) { return nil, false }

type noSingle struct{ noMulti }

func (noMulti) ToStudy() (Study, bool) { return nil, false }

// MultiVarStudy represents a study that can accept multiple variables and can return multiple values
// for example MACD or VWAP
type MultiVarStudy interface {
	Study
	// UpdateAll same as `Study.Update`, however will return multiple values
	// for example MACD or VWAP with Bands
	UpdateAll(values ...Decimal) []Decimal

	// LenAll returns the different periods of all underlying studies, if any
	LenAll() []int

	// ToStudy can be used to convert the Multi to a normal study if supported
	ToStudy() (Study, bool)
}

// ApplyStudy applies the given study to the input(s) and returns the result(s)
func ApplyStudy(s Study, tas ...*TA) []*TA {
	if sws, ok := s.(StudyWithSetup); ok {
		return sws.Setup(tas...)
	}
	out := make([]*TA, len(tas))
	ln := tas[0].Len()
	for i := 0; i < len(tas); i++ {
		out[i] = NewSize(ln, true)
	}
	for i := 0; i < ln; i++ {
		for j := 0; j < len(tas); j++ {
			out[j].Append(tas[j].Get(i))
		}
	}
	return out
}

// ApplyMultiVarStudy applies the given study to input(s) and returns the result(s)
func ApplyMultiVarStudy(s MultiVarStudy, tas ...*TA) []*TA {
	if sws, ok := s.(StudyWithSetup); ok {
		return sws.Setup(tas...)
	}

	out := make([]*TA, len(tas))
	vals := make([]Decimal, len(tas))
	ln := tas[0].Len()
	for i := 0; i < len(tas); i++ {
		out[i] = NewSize(ln, true)
	}
	for i := 0; i < ln; i++ {
		for j := 0; j < len(tas); j++ {
			vals[j] = tas[j].Get(i)
		}
		for j, v := range s.UpdateAll(vals...) {
			out[j].Append(v)
		}
	}

	return out
}

// RSI - Relative Strength Index
func RSI(period int) Study {
	checkPeriod(period, 2)
	return &rsi{
		period: period,
		per:    1 / Decimal(period),
	}
}

// RSIExt - Relative Strength Index using a different moving average func
func RSIExt(ma MovingAverage) Study {
	period := ma.Len()
	checkPeriod(period, 2)
	return &rsi{
		ext:    ma,
		period: period,
		per:    1 / Decimal(period),
	}
}

var _ Study = (*rsi)(nil)

type rsi struct {
	noMulti
	ext        MovingAverage
	prev       Decimal
	smoothUp   Decimal
	smoothDown Decimal
	per        Decimal
	period     int
	idx        int
}

func (l *rsi) Update(vs ...Decimal) Decimal {
	for _, v := range vs {
		if l.ext != nil {
			v = l.ext.Update(v)
		}
		prev := l.prev
		l.prev = v

		if l.idx > l.period {
			var up, down Decimal
			if v > prev {
				up = v - prev
			}
			if v < prev {
				down = prev - v
			}
			l.smoothUp = (up-l.smoothUp)*l.per + l.smoothUp
			l.smoothDown = (down-l.smoothDown)*l.per + l.smoothDown
			continue
		}

		if l.idx == 0 {
			l.idx++
		} else {
			if v > prev {
				l.smoothUp += v - prev
			}
			if v < prev {
				l.smoothDown += prev - v
			}

			if l.idx == l.period {
				l.smoothUp /= Decimal(l.period)
				l.smoothDown /= Decimal(l.period)
			}

			l.idx++
		}
	}

	upDown := l.smoothUp + l.smoothDown
	if upDown == 0 {
		return l.prev
	}

	return 100 * (l.smoothUp / upDown)
}

func (l *rsi) Len() int { return l.period }

// MACD - Moving Average Convergence/Divergence, using EMA for all periods
// alias for MACDExt(fastPeriod, slowPeriod, signalPeriod, EMA)
func MACD(fastPeriod, slowPeriod, signalPeriod int) MultiVarStudy {
	return MACDExt(fastPeriod, slowPeriod, signalPeriod, EMA)
}

// MACDExt - MACD using the specified MA func for all periods
// alias for MACDMulti(ma(fastPeriod), ma(slowPeriod), ma(signalPeriod))
func MACDExt(fastPeriod, slowPeriod, signalPeriod int, ma MovingAverageFunc) MultiVarStudy {
	return MACDMulti(ma(fastPeriod), ma(slowPeriod), ma(signalPeriod))
}

// MACDMulti - MACD that supports different MA funcs for each period
// returns a multi study, however it can work as a normal Study
// Update will return the diff value
func MACDMulti(fast, slow, signal MovingAverage) MultiVarStudy {
	if slow.Len() < fast.Len() {
		slow, fast = fast, slow
	}
	return &macd{
		slow:   slow,
		fast:   fast,
		signal: signal,
		prev:   math.MaxFloat64,
	}
}

var (
	_ Study         = (*macd)(nil)
	_ MultiVarStudy = (*macd)(nil)
)

type macd struct {
	slow, fast, signal MovingAverage

	prev Decimal
}

func (l *macd) Setup(ds ...*TA) []*TA {
	d0ln := ds[0].Len()
	macd := NewSize(d0ln, false)
	signal := NewSize(d0ln, false)
	hist := NewSize(d0ln, false)
	for _, d := range ds {
		for i := 0; i < d.Len(); i++ {
			v := d.Get(i)
			vs := l.UpdateAll(v)
			macd.Set(i, vs[0])
			signal.Set(i, vs[1])
			hist.Set(i, vs[2])
		}
	}
	macd = macd.Slice(-l.signal.Len(), 0)
	signal = signal.Slice(-l.signal.Len(), 0)
	hist = hist.Slice(-l.signal.Len(), 0)
	return []*TA{macd, signal, hist}
}

func (l *macd) Update(vs ...Decimal) Decimal {
	var fast, slow, macd, sig Decimal
	for _, v := range vs {
		fast = l.fast.Update(v)
		slow = l.slow.Update(v)
		macd = fast - slow
		sig = l.signal.Update(macd)
	}
	return macd - sig
}

func (l *macd) UpdateAll(vs ...Decimal) []Decimal {
	var fast, slow, macd, sig Decimal
	for _, v := range vs {
		fast = l.fast.Update(v)
		slow = l.slow.Update(v)
		macd = fast - slow
		sig = l.signal.Update(macd)
	}
	return []Decimal{macd, sig, macd - sig}
}

func (l *macd) Len() int { return l.signal.Len() }
func (l *macd) LenAll() []int {
	return []int{l.slow.Len(), l.fast.Len(), l.signal.Len()}
}

func (l *macd) ToMulti() (MultiVarStudy, bool) { return l, true }
func (l *macd) ToStudy() (Study, bool)         { return l, true }

// VWAP - Volume Weighted Average Price
// alias for VWAPBands(period, -period)
func VWAP(period int) Study {
	d := Decimal(period)
	return VWAPBands(d, -d)
}

// VWAPBands - Volume Weighted Average Price with upper and lower bands
// Update/UpdateAll expects 2 values, the volume and price, it will panic otherwise
// Update returns VWAP
// UpdateAll returns [VWAP, UPPER, LOWER]
func VWAPBands(up, down Decimal) MultiVarStudy {
	if up < down {
		up, down = down, up
	}
	period := (up - down) / 2
	checkPeriod(int(period), 2)
	return &vwap{
		std:  newVar(int(period), runStd),
		up:   up,
		down: down,
	}
}

var (
	_ Study         = (*macd)(nil)
	_ MultiVarStudy = (*macd)(nil)
)

type vwap struct {
	std   *variance
	up    Decimal
	down  Decimal
	sum   Decimal
	total Decimal
}

func (l *vwap) Setup(ds ...*TA) []*TA {
	if len(ds) != 2 {
		panic("vwap: must pass volume and price")
	}
	vol, price := ds[0], ds[1]
	if vol.Len() != price.Len() {
		panic("vwap: vol.Len() != price.Len()")
	}

	vw := NewCapped(l.std.Len())
	up := NewCapped(l.std.Len())
	dn := NewCapped(l.std.Len())
	for i := 0; i < vol.Len(); i++ {
		v, p := vol.Get(i), price.Get(i)
		vud := l.UpdateAll(v, p)
		vw.Append(vud[0])
		up.Append(vud[1])
		dn.Append(vud[2])

	}
	return []*TA{vw, up, dn}
}

func (l *vwap) Update(vs ...Decimal) Decimal {
	return l.UpdateAll(vs...)[0]
}

func (l *vwap) UpdateAll(vs ...Decimal) []Decimal {
	if len(vs) != 2 {
		panic("vwap: must provide volume and price")
	}
	vol, price := vs[0], vs[1]
	l.sum += vol * price
	l.total += vol

	var (
		v    = l.sum / l.total
		d    = l.std.Update(v)
		up   = v + d*l.up
		down Decimal
	)

	if l.up != l.down {
		down = v + d*l.down
	}

	return []Decimal{v, up, down}
}

func (l *vwap) Len() int      { return l.std.Len() }
func (l *vwap) LenAll() []int { return []int{l.std.Len()} }

func (l *vwap) ToMulti() (MultiVarStudy, bool) { return l, true }
func (l *vwap) ToStudy() (Study, bool)         { return l, true }
