package ta // import "go.oneofone.dev/ta"

// ref: https://github.com/TulipCharts/tulipindicators/tree/master/indicators

// MovingAverageFunc defines a function that returns am updatable moving average for the given period

type MovingAverageFunc func(period int) Live

func (sl TA) MovingAverage(fn MovingAverageFunc, period int) (TA, Live) {
	ma := fn(period)
	return ma.Setup(sl), ma
}

// SMA - Simple Moving Average
func (sl TA) SMA(period int) (TA, Live) {
	return sl.MovingAverage(SMA, period)
}

// SMA - Simple Moving Average
func SMA(period int) Live {
	checkPeriod(period, 2)
	return &liveSMA{
		data:   make(TA, period),
		period: period,
	}
}

type liveSMA struct {
	data   TA
	sum    F
	period int
	idx    int
	count  int
}

func (l *liveSMA) Setup(d TA) TA {
	return d.Apply(l.Update, false).Slice(-l.period+1, 0)
}

func (l *liveSMA) Update(v F) F {
	l.idx = (l.idx + 1) % l.period
	// prev := l.data[l.idx]
	prev := l.data[l.idx]
	l.data[l.idx] = v

	if l.count < l.period {
		l.count++
	}

	l.sum = l.sum - prev + v
	return l.sum / F(l.count)
}

func (l *liveSMA) Len() int { return l.period }

func (l *liveSMA) Clone() Live {
	cp := *l
	cp.data = l.data.Copy()
	return &cp
}

// EMA - Exponential Moving Average
func (sl TA) EMA(period int) (TA, Live) {
	return sl.MovingAverage(EMA, period)
}

// EMA - Exponential Moving Average
// An alias for CustomEMA(period, 2 / (period+1))
func EMA(period int) Live {
	return CustomEMA(period, 0)
}

// CustomEMA - returns an updatable EMA with the given k
func CustomEMA(period int, k F) Live {
	checkPeriod(period, 2)
	if k == 0 {
		k = F(2 / float64(period+1))
	}
	return &liveEMA{
		k:      k,
		period: period,
	}
}

type liveEMA struct {
	k      F
	prevMA F
	period int
	idx    int
	set    bool
}

func (l *liveEMA) Setup(d TA) TA {
	l.set = true
	l.prevMA = d.Slice(0, l.period).Avg()
	d = d.Slice(l.period, 0).Apply(l.Update, false)
	return d
}

func (l *liveEMA) Update(v F) F {
	if l.set {
		l.prevMA = v.Sub(l.prevMA).Mul(l.k).Add(l.prevMA)
		return l.prevMA
	}

	l.prevMA += v
	if l.idx++; l.idx == l.period {
		l.set = true
		l.prevMA = l.prevMA / F(l.period)
	}
	return l.prevMA
}

func (l *liveEMA) Len() int { return l.period }

func (l *liveEMA) Clone() Live {
	cp := *l
	return &cp
}

func (l *liveEMA) copy() liveEMA {
	return *l
}

// WMA - Exponential Moving Average
func (sl TA) WMA(period int) (TA, Live) {
	return sl.MovingAverage(WMA, period)
}

// WMA - Exponential Moving Average
// An alias for CustomWMA(period, (period * (period + 1)) >> 1)
func WMA(period int) Live {
	w := F((period * (period + 1)) >> 1)
	return CustomWMA(period, w)
}

// CustomWMA returns an updatable WMA with the given weight
func CustomWMA(period int, weight F) Live {
	checkPeriod(period, 2)
	return &liveWMA{
		data:   make(TA, period),
		weight: weight,
		period: period,
	}
}

type liveWMA struct {
	data        TA
	weight      F
	sum         F
	weightedSum F
	idx         int
	period      int
	set         bool
}

func (l *liveWMA) Setup(d TA) TA {
	l.set = true
	var sum, wsum F
	for i := 0; i < l.period-1; i++ {
		v := d[i]
		wsum += v * F(i+1)
		sum += v
		l.data[i] = v
	}
	l.sum, l.weightedSum = sum, wsum
	l.idx = l.period - 2
	d = d.Slice(l.period-1, 0).Apply(l.Update, false)
	return d
}

func (l *liveWMA) Update(v F) F {
	if !l.set {
		l.data[l.idx] = v
		if l.idx < l.period-1 {
			l.idx++
			l.weightedSum += v * F(l.idx)
			l.sum += v
			return l.weightedSum / (l.weight * F(l.idx))
		}
		l.idx = l.period - 2
		l.set = true
	}
	l.idx = (l.idx + 1) % l.period
	l.data[l.idx] = v
	l.weightedSum += v * F(l.period)
	l.sum += v
	rv := l.weightedSum / l.weight
	l.weightedSum -= l.sum

	pidx := (l.idx + 1) % l.period
	l.sum -= l.data[pidx]
	return rv
}

func (l *liveWMA) Len() int { return l.period }

func (l *liveWMA) Clone() Live {
	cp := *l
	cp.data = l.data.Copy()
	return &cp
}

// DEMA - Double Exponential Moving Average
func (sl TA) DEMA(period int) (TA, Live) {
	return sl.MovingAverage(DEMA, period)
}

// DEMA - Double Exponential Moving Average
func DEMA(period int) Live {
	checkPeriod(period, 2)
	e1 := EMA(period).(*liveEMA)
	return &liveDEMA{
		e1:     *e1,
		period: period,
	}
}

type liveDEMA struct {
	e1, e2 liveEMA
	period int
	idx    int
}

func (l *liveDEMA) Setup(d TA) TA { return d.Apply(l.Update, false).Slice(-l.period, 0) }

func (l *liveDEMA) Update(v F) F {
	e1 := l.e1.Update(v)
	if l.idx < l.period {
		if l.idx++; l.idx == l.period {
			l.e2 = l.e1.copy()
		}
		return v
	}
	e2 := l.e2.Update(e1)
	return e1*2 - e2
}

func (l *liveDEMA) Len() int { return l.period }

func (l *liveDEMA) Clone() Live {
	cp := *l
	cp.e1 = cp.e1.copy()
	cp.e2 = cp.e2.copy()
	return &cp
}

// TEMA - Triple Exponential Moving Average
func (sl TA) TEMA(period int) (TA, Live) {
	return sl.MovingAverage(TEMA, period)
}

// TEMA - Triple Exponential Moving Average
func TEMA(period int) Live {
	checkPeriod(period, 2)
	e1 := EMA(period).(*liveEMA)
	return &liveTEMA{
		e1:     *e1,
		period: period,
	}
}

type liveTEMA struct {
	e1, e2, e3 liveEMA
	period     int
	idx        int
	max2       int
	max3       int
}

func (l *liveTEMA) Setup(d TA) TA { return d.Apply(l.Update, false).Slice(-l.period, 0) }

func (l *liveTEMA) Update(v F) F {
	e1 := l.e1.Update(v)
	if l.idx < l.period {
		if l.idx++; l.idx == l.period {
			l.e2 = l.e1.copy()
			l.e3 = l.e1.copy()
		}
		return v
	}
	e2 := l.e2.Update(e1)
	e3 := l.e3.Update(e2)
	return 3*e1 - 3*e2 + e3
}

func (l *liveTEMA) Len() int { return l.period }

func (l *liveTEMA) Clone() Live {
	cp := *l
	cp.e1 = *(l.e1.Clone().(*liveEMA))
	cp.e2 = *(l.e2.Clone().(*liveEMA))
	return &cp
}

// TODO:
// - Trima
// - KAMA
// - MAMA/FAMA
// - T3
