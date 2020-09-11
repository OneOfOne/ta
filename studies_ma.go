package ta

type MovingAverage interface {
	Setup(d *TA) *TA
	Update(v Decimal) Decimal
	Len() int
	ma()
}

type implMA struct{}

func (implMA) ma() {}

// MovingAverageFunc defines a function that returns am updatable moving average for the given period
type MovingAverageFunc func(period int) MovingAverage

func (ta *TA) MovingAverage(fn MovingAverageFunc, period int) (*TA, MovingAverage) {
	ma := fn(period)
	return ma.Setup(ta), ma
}

// SMA - Simple Moving Average
func (ta *TA) SMA(period int) (*TA, MovingAverage) {
	return ta.MovingAverage(SMA, period)
}

// SMA - Simple Moving Average
func SMA(period int) MovingAverage {
	checkPeriod(period, 2)
	return &sma{
		data:   NewCapped(period),
		period: period,
	}
}

type sma struct {
	implMA
	data   *TA
	idx    int
	sum    Decimal
	period int
	count  int
}

func (l *sma) Setup(d *TA) *TA {
	return d.Map(l.Update, false).Slice(-l.period, 0)
}

func (l *sma) Update(v Decimal) Decimal {
	prev := l.data.Update(v)
	if l.count < l.period {
		l.count++
	}

	l.sum = l.sum - prev + v
	return l.sum / Decimal(l.count)
}

func (l *sma) Len() int { return l.period }

// EMA - Exponential Moving Average
func (ta *TA) EMA(period int) (*TA, MovingAverage) {
	return ta.MovingAverage(EMA, period)
}

// EMA - Exponential Moving Average
// An alias for CustomEMA(period, 2 / (period+1))
func EMA(period int) MovingAverage {
	return CustomEMA(period, 0)
}

// CustomEMA - returns an updatable EMA with the given k
func CustomEMA(period int, k Decimal) MovingAverage {
	checkPeriod(period, 2)
	if k == 0 {
		k = Decimal(2 / float64(period+1))
	}
	return &ema{
		k:      k,
		period: period,
	}
}

type ema struct {
	implMA
	k      Decimal
	prevMA Decimal
	period int
	idx    int
	set    bool
}

func (l *ema) Setup(d *TA) *TA {
	l.set = true
	l.prevMA = d.Slice(0, l.period).Avg()
	d = d.Slice(l.period, 0).Map(l.Update, false)
	return d
}

func (l *ema) Update(v Decimal) Decimal {
	if l.set {
		l.prevMA = v.Sub(l.prevMA).Mul(l.k).Add(l.prevMA)
		return l.prevMA
	}

	l.prevMA += v
	if l.idx++; l.idx == l.period {
		l.set = true
		l.prevMA = l.prevMA / Decimal(l.period)
	}
	return l.prevMA
}

func (l *ema) Len() int { return l.period }

func (l *ema) copy() ema {
	return *l
}

// WMA - Exponential Moving Average
func (ta *TA) WMA(period int) (*TA, MovingAverage) {
	return ta.MovingAverage(WMA, period)
}

// WMA - Exponential Moving Average
// An alias for CustomWMA(period, (period * (period + 1)) >> 1)
func WMA(period int) MovingAverage {
	w := Decimal((period * (period + 1)) >> 1)
	return CustomWMA(period, w)
}

// CustomWMA returns an updatable WMA with the given weight
func CustomWMA(period int, weight Decimal) MovingAverage {
	checkPeriod(period, 2)
	return &wma{
		data:   NewSize(period, false),
		weight: weight,
		period: period,
	}
}

type wma struct {
	implMA
	data        *TA
	weight      Decimal
	sum         Decimal
	weightedSum Decimal
	idx         int
	period      int
	set         bool
}

func (l *wma) Setup(d *TA) *TA {
	l.set = true
	var sum, wsum Decimal
	for i := 0; i < l.period-1; i++ {
		v := d.Get(i)
		wsum += v * Decimal(i+1)
		sum += v
		l.data.Set(i, v)
	}
	l.sum, l.weightedSum = sum, wsum
	l.idx = l.period - 2
	d = d.Slice(l.period-1, 0).Map(l.Update, false)
	return d
}

func (l *wma) Update(v Decimal) Decimal {
	if !l.set {
		l.data.Set(l.idx, v)
		if l.idx < l.period-1 {
			l.idx++
			l.weightedSum += v * Decimal(l.idx)
			l.sum += v
			return l.weightedSum / (l.weight * Decimal(l.idx))
		}
		l.idx = l.period - 2
		l.set = true
	}
	l.idx = (l.idx + 1) % l.period
	l.data.Set(l.idx, v)
	l.weightedSum += v * Decimal(l.period)
	l.sum += v
	rv := l.weightedSum / l.weight
	l.weightedSum -= l.sum

	pidx := (l.idx + 1) % l.period
	l.sum -= l.data.Get(pidx)
	return rv
}

func (l *wma) Len() int { return l.period }

// DEMA - Double Exponential Moving Average
func (ta *TA) DEMA(period int) (*TA, MovingAverage) {
	return ta.MovingAverage(DEMA, period)
}

// DEMA - Double Exponential Moving Average
func DEMA(period int) MovingAverage {
	checkPeriod(period, 2)
	e1 := EMA(period).(*ema)
	return &dema{
		e1:     *e1,
		period: period,
	}
}

type dema struct {
	implMA
	e1, e2 ema
	period int
	idx    int
}

func (l *dema) Setup(d *TA) *TA { return d.Map(l.Update, false).Slice(-l.period, 0) }

func (l *dema) Update(v Decimal) Decimal {
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

func (l *dema) Len() int { return l.period }

// TEMA - Triple Exponential Moving Average
func (ta *TA) TEMA(period int) (*TA, MovingAverage) {
	return ta.MovingAverage(TEMA, period)
}

// TEMA - Triple Exponential Moving Average
func TEMA(period int) MovingAverage {
	checkPeriod(period, 2)
	e1 := EMA(period).(*ema)
	return &tema{
		e1:     *e1,
		period: period,
	}
}

type tema struct {
	implMA
	e1, e2, e3 ema
	period     int
	idx        int
	max2       int
	max3       int
}

func (l *tema) Setup(d *TA) *TA { return d.Map(l.Update, false).Slice(-l.period, 0) }

func (l *tema) Update(v Decimal) Decimal {
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

func (l *tema) Len() int { return l.period }

// TODO:
// - Trima
// - KAMA
// - MAMA/FAMA
// - T3
