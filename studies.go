package ta

import "math"

type Study interface {
	Setup(d *TA) *TA
	Update(v Decimal) Decimal
	Len() int
	Clone() Study
}

// RSI - Relative Strength Index
func (ta *TA) RSI(period int) (*TA, Study) {
	return ta.MovingAverage(RSI, period)
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
func RSIExt(ma Study) Study {
	period := ma.Len()
	checkPeriod(period, 2)
	return &liveRSI{
		ext:    ma,
		period: period,
		per:    1 / Decimal(period),
	}
}

type liveRSI struct {
	ext        Study
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

func (l *liveRSI) Clone() Study {
	cp := *l
	return &cp
}

type MACDStudy interface {
	Setup(*TA) (*TA, *TA, *TA)
	Update(v Decimal) (Decimal, Decimal, Decimal)
	Len() (int, int, int)
	Clone() MACDStudy
}

// MACD - Moving Average Convergence/Divergence, using EMA
func (ta *TA) MACD(fastPeriod, slowPeriod, signalPeriod int) (macd, signal, hist *TA, ma MACDStudy) {
	ma = MACD(fastPeriod, slowPeriod, signalPeriod)
	macd, signal, hist = ma.Setup(ta)
	return
	// return sl.MovingAverage(MACD, period)
}

// MACDExt - Moving Average Convergence/Divergence using custom MA functions
func (ta *TA) MACDExt(fastMA, slowMA, signalMA Study) (macd, signal, hist *TA, ma MACDStudy) {
	ma = MACDExt(fastMA, slowMA, signalMA)
	macd, signal, hist = ma.Setup(ta)
	return
	// return sl.MovingAverage(MACD, period)
}

// MACD - Moving Average Convergence/Divergence, usign EMA
func MACD(fastPeriod, slowPeriod, signalPeriod int) MACDStudy {
	return MACDExt(EMA(fastPeriod), EMA(slowPeriod), EMA(signalPeriod))
}

// MACDExt - Moving Average Convergence/Divergence using custom MA functions
func MACDExt(fast, slow, signal Study) MACDStudy {
	if slow.Len() < fast.Len() {
		slow, fast = fast, slow
	}
	return &liveMACD{
		slow:   slow,
		fast:   fast,
		signal: signal,
		prev:   math.MaxFloat64,
	}
}

type liveMACD struct {
	slow, fast, signal Study

	prev Decimal
}

func (l *liveMACD) Setup(d *TA) (macd, signal, hist *TA) {
	macd = NewSize(d.Len(), false)
	signal = NewSize(d.Len(), false)
	hist = NewSize(d.Len(), false)
	for i := 0; i < d.Len(); i++ {
		v := d.At(i)
		a, b, c := l.Update(v)
		macd.SetAt(i, a)
		signal.SetAt(i, b)
		hist.SetAt(i, c)
	}
	macd = macd.Slice(-l.signal.Len(), 0)
	signal = signal.Slice(-l.signal.Len(), 0)
	hist = hist.Slice(-l.signal.Len(), 0)
	return
}

func (l *liveMACD) Update(v Decimal) (Decimal, Decimal, Decimal) {
	fast := l.fast.Update(v)
	slow := l.slow.Update(v)
	macd := fast - slow
	sig := l.signal.Update(macd)
	return macd, sig, macd - sig
}

func (l *liveMACD) Len() (int, int, int) {
	return l.slow.Len(), l.fast.Len(), l.signal.Len()
}

func (l *liveMACD) Clone() MACDStudy {
	panic("x")
}
