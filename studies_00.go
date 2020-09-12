package ta

import (
	"math"
)

type Study interface {
	Update(v ...Decimal) Decimal
	Len() int
	Multi() (Multi, bool)
}

type StudyWithSetup interface {
	Study
	Setup(...*TA) []*TA
}

type noMulti struct{}

func (noMulti) Multi() (Multi, bool) { return nil, false }

type noSingle struct{ noMulti }

func (noMulti) Single() (Study, bool) { return nil, false }

// Multi represents a study that accepts multiple values and returns multiple values, for example, MACD
type Multi interface {
	Study
	UpdateAll(...Decimal) []Decimal
	LenAll() []int
	Single() (Study, bool)
}

func ApplyStudy(s Study, ta *TA) *TA {
	if sws, ok := s.(StudyWithSetup); ok {
		return sws.Setup(ta)[0]
	}
	period := s.Len()
	ta = ta.Map(func(d Decimal) Decimal { return s.Update(d) }, false)
	return ta.Slice(-period, 0)
}

func ApplyMultiStudy(s Multi, tas ...*TA) []*TA {
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
func MACD(fastPeriod, slowPeriod, signalPeriod int) Multi {
	return MACDExt(fastPeriod, slowPeriod, signalPeriod, EMA)
}

// MACDExt - Moving Average Convergence/Divergence, using a custom MA for all periods
func MACDExt(fastPeriod, slowPeriod, signalPeriod int, ma MovingAverageFunc) Multi {
	if ma == nil {
		panic("ma == nil")
	}
	return MACDMulti(ma(fastPeriod), ma(slowPeriod), ma(signalPeriod))
}

// MACDMulti - Moving Average Convergence/Divergence using a custom MA functions for each period
func MACDMulti(fast, slow, signal MovingAverage) Multi {
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
	_ Study = (*macd)(nil)
	_ Multi = (*macd)(nil)
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
	return l.UpdateAll(vs...)[2]
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

func (l *macd) Multi() (Multi, bool)  { return l, true }
func (l *macd) Single() (Study, bool) { return l, true }

func VWAP(period int) Multi {
	p := Decimal(period)
	return VWAPBands(p, -p)
}

func VWAPBands(up, down Decimal) Multi {
	period := (up - down).Abs() / 2
	return &vwap{
		std:  newVar(int(period), runStd),
		up:   up,
		down: down,
	}
}

var (
	_ Study = (*macd)(nil)
	_ Multi = (*macd)(nil)
)

type vwap struct {
	noSingle
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
