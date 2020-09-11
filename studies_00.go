package ta

import (
	"math"
)

type Study interface {
	Setup(d *TA) *TA
	Update(v Decimal) Decimal
	Len() int
}

// StudyMulti represents a study that accepts multiple values and returns multiple values, for example, MACD
type StudyMulti interface {
	Setup(...*TA) []*TA
	Update(v ...Decimal) []Decimal
	Len() []int
}

// RSI - Relative Strength Index
func (ta *TA) RSI(period int) (*TA, Study) {
	s := RSI(period)
	ta = s.Setup(ta)
	return ta, s
}

// RSI - Relative Strength Index
func RSI(period int) Study {
	checkPeriod(period, 2)
	return &liveRSI{
		period: period,
		per:    1 / Decimal(period),
	}
}

// RSIExt - Relative Strength Index using a different moving average func
func RSIExt(ma MovingAverage) Study {
	period := ma.Len()
	checkPeriod(period, 2)
	return &liveRSI{
		ext:    ma,
		period: period,
		per:    1 / Decimal(period),
	}
}

type liveRSI struct {
	ext        MovingAverage
	prev       Decimal
	smoothUp   Decimal
	smoothDown Decimal
	per        Decimal
	period     int
	idx        int
}

func (l *liveRSI) Setup(d *TA) *TA {
	return d.Map(l.Update, false).Slice(-l.period+1, 0)
}

func (l *liveRSI) Update(v Decimal) Decimal {
	if l.ext != nil {
		v = l.ext.Update(v)
	}
	if l.idx == 0 {
		l.idx++
		l.prev = v
		return v
	} else if l.idx <= l.period {
		if v > l.prev {
			l.smoothUp += v - l.prev
		}
		if v < l.prev {
			l.smoothDown += l.prev - v
		}

		if l.idx == l.period {
			l.smoothUp /= Decimal(l.period)
			l.smoothDown /= Decimal(l.period)
		}

		l.idx++
	} else {
		var up, down Decimal
		if v > l.prev {
			up = v - l.prev
		}
		if v < l.prev {
			down = l.prev - v
		}
		l.smoothUp = (up-l.smoothUp)*l.per + l.smoothUp
		l.smoothDown = (down-l.smoothDown)*l.per + l.smoothDown
	}

	l.prev = v
	return 100 * (l.smoothUp / (l.smoothUp + l.smoothDown))
}

func (l *liveRSI) Len() int { return l.period }

// MACD - Moving Average Convergence/Divergence, using EMA
func (ta *TA) MACD(fastPeriod, slowPeriod, signalPeriod int) (macd, signal, hist *TA, _ StudyMulti) {
	ma := MACD(fastPeriod, slowPeriod, signalPeriod)
	out := ma.Setup(ta)
	return out[0], out[1], out[2], ma
}

// MACDExt - Moving Average Convergence/Divergence, using a custom MA for all periods
func (ta *TA) MACDExt(fastPeriod, slowPeriod, signalPeriod int, maf MovingAverageFunc) (macd, signal, hist *TA, _ StudyMulti) {
	ma := MACDExt(fastPeriod, slowPeriod, signalPeriod, maf)
	out := ma.Setup(ta)
	return out[0], out[1], out[2], ma
}

// MACDMulti - Moving Average Convergence/Divergence using a custom MA functions for each period
func (ta *TA) MACDMulti(fastMA, slowMA, signalMA MovingAverage) (macd, signal, hist *TA, _ StudyMulti) {
	ma := MACDMulti(fastMA, slowMA, signalMA)
	out := ma.Setup(ta)
	return out[0], out[1], out[2], ma
}

// MACD - Moving Average Convergence/Divergence, using EMA for all periods
func MACD(fastPeriod, slowPeriod, signalPeriod int) StudyMulti {
	return MACDExt(fastPeriod, slowPeriod, signalPeriod, EMA)
}

// MACDExt - Moving Average Convergence/Divergence, using a custom MA for all periods
func MACDExt(fastPeriod, slowPeriod, signalPeriod int, ma MovingAverageFunc) StudyMulti {
	if ma == nil {
		panic("ma == nil")
	}
	return MACDMulti(ma(fastPeriod), ma(slowPeriod), ma(signalPeriod))
}

// MACDMulti - Moving Average Convergence/Divergence using a custom MA functions for each period
func MACDMulti(fast, slow, signal MovingAverage) StudyMulti {
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
			vs := l.Update(v)
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

func (l *macd) Update(vs ...Decimal) []Decimal {
	var fast, slow, macd, sig Decimal
	for _, v := range vs {
		fast = l.fast.Update(v)
		slow = l.slow.Update(v)
		macd = fast - slow
		sig = l.signal.Update(macd)
	}
	return []Decimal{macd, sig, macd - sig}
}

func (l *macd) Len() []int {
	return []int{l.slow.Len(), l.fast.Len(), l.signal.Len()}
}

func VWAP(period int) StudyMulti {
	p := Decimal(period)
	return VWAPBands(p, -p)
}

func VWAPBands(up, down Decimal) StudyMulti {
	period := (up - down).Abs() / 2
	return &vwap{
		std:  newVar(int(period), runStd),
		up:   up,
		down: down,
	}
}

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
		vud := l.Update(v, p)
		vw.Append(vud[0])
		up.Append(vud[1])
		dn.Append(vud[2])

	}
	return []*TA{vw, up, dn}
}

func (l *vwap) Update(vs ...Decimal) []Decimal {
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

func (l *vwap) Len() []int {
	return []int{l.std.Len()}
}
