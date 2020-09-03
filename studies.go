package ta

import "math"

// RSI - Simple Moving Average
func (sl TA) RSI(period int) (TA, Live) {
	return sl.MovingAverage(RSI, period)
}

// RSI - Simple Moving Average
func RSI(period int) Live {
	checkPeriod(period, 2)
	return &liveRSI{
		period: period,
		per:    1 / F(period),
	}
}

type liveRSI struct {
	prev       F
	smoothUp   F
	smoothDown F
	per        F
	period     int
	idx        int
}

func (l *liveRSI) Setup(d TA) TA {
	return d.Apply(l.Update, false).Slice(-l.period+1, 0)
}

func (l *liveRSI) Update(v F) F {
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
			l.smoothUp /= F(l.period)
			l.smoothDown /= F(l.period)
		}

		l.idx++
	} else {
		var up, down F
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

func (l *liveRSI) Clone() Live {
	cp := *l
	return &cp
}

type LiveMACD interface {
	Setup(TA) (TA, TA, TA)
	Update(v F) (F, F, F)
	Len() (int, int, int)
	Clone() LiveMACD
}

// MACD - Moving Average Convergence/Divergence, using EMA
func (sl TA) MACD(fastPeriod, slowPeriod, signalPeriod int) (macd, signal, hist TA, ma LiveMACD) {
	ma = MACD(fastPeriod, slowPeriod, signalPeriod)
	macd, signal, hist = ma.Setup(sl)
	return
	// return sl.MovingAverage(MACD, period)
}

// MACDExt - Moving Average Convergence/Divergence using custom MA functions
func (sl TA) MACDExt(fastMA, slowMA, signalMA Live) (macd, signal, hist TA, ma LiveMACD) {
	ma = MACDExt(fastMA, slowMA, signalMA)
	macd, signal, hist = ma.Setup(sl)
	return
	// return sl.MovingAverage(MACD, period)
}

// MACD - Moving Average Convergence/Divergence, usign EMA
func MACD(fastPeriod, slowPeriod, signalPeriod int) LiveMACD {
	return MACDExt(EMA(fastPeriod), EMA(slowPeriod), EMA(signalPeriod))
}

// MACDExt - Moving Average Convergence/Divergence using custom MA functions
func MACDExt(fast, slow, signal Live) LiveMACD {
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
	slow, fast, signal Live

	prev F
}

func (l *liveMACD) Setup(d TA) (macd, signal, hist TA) {
	macd = make(TA, d.Len())
	signal = make(TA, d.Len())
	hist = make(TA, d.Len())
	for i, v := range d {
		macd[i], signal[i], hist[i] = l.Update(v)
	}
	macd = macd.Slice(-l.signal.Len(), 0)
	signal = signal.Slice(-l.signal.Len(), 0)
	hist = hist.Slice(-l.signal.Len(), 0)
	return
}

func (l *liveMACD) Update(v F) (F, F, F) {
	fast := l.fast.Update(v)
	slow := l.slow.Update(v)
	macd := fast - slow
	sig := l.signal.Update(macd)
	return macd, sig, macd - sig
}

func (l *liveMACD) Len() (int, int, int) {
	return l.slow.Len(), l.fast.Len(), l.signal.Len()
}

func (l *liveMACD) Clone() LiveMACD {
	panic("x")
}
