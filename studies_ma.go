package ta

type MovingAverage interface {
	Study
	ma()
}

type implMA struct {
	noMulti
}

func (implMA) ma() {}

// MovingAverageFunc defines a function that returns am updatable moving average for the given period
type MovingAverageFunc func(period int) MovingAverage

func (ta *TA) MovingAverage(fn MovingAverageFunc, period int) (*TA, MovingAverage) {
	ma := fn(period)
	ta = ApplyStudy(ma, ta)
	return ta, ma
}

// SMA - Simple Moving Average
func SMA(period int) MovingAverage {
	checkPeriod(period, 2)
	return &sma{
		data:   NewCapped(period),
		period: period,
	}
}

var _ Study = (*sma)(nil)

type sma struct {
	implMA
	data   *TA
	idx    int
	sum    Decimal
	period int
	count  int
}

func (l *sma) Update(vs ...Decimal) Decimal {
	for _, v := range vs {
		prev := l.data.Update(v)
		if l.count < l.period {
			l.count++
		}

		l.sum = l.sum - prev + v
	}
	return l.sum / Decimal(l.count)
}

func (l *sma) Len() int { return l.period }

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

var _ Study = (*ema)(nil)

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
	d = d.Slice(l.period, 0).Map(func(d Decimal) Decimal { return l.Update(d) }, false)
	return d
}

func (l *ema) Update(vs ...Decimal) Decimal {
	for _, v := range vs {
		if l.set {
			l.prevMA = v.Sub(l.prevMA).Mul(l.k).Add(l.prevMA)
			return l.prevMA
		}

		l.prevMA += v
		if l.idx++; l.idx == l.period {
			l.set = true
			l.prevMA = l.prevMA / Decimal(l.period)
		}
	}
	return l.prevMA
}

func (l *ema) Len() int { return l.period }

func (l *ema) copy() ema {
	return *l
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

var _ Study = (*wma)(nil)

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
	d = d.Slice(l.period-1, 0).Map(func(d Decimal) Decimal { return l.Update(d) }, false)
	return d
}

func (l *wma) Update(vs ...Decimal) (rv Decimal) {
	for _, v := range vs {
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
		rv = l.weightedSum / l.weight
		l.weightedSum -= l.sum

		pidx := (l.idx + 1) % l.period
		l.sum -= l.data.Get(pidx)
	}
	return rv
}

func (l *wma) Len() int { return l.period }

// DEMA - Double Exponential Moving Average
func DEMA(period int) MovingAverage {
	return DoubleMA(period, EMA)
}

// DoubleMA - Double Moving Average
func DoubleMA(period int, ma MovingAverageFunc) MovingAverage {
	checkPeriod(period, 2)
	return &dxma{
		e1: ma(period),
		e2: ma(period),
	}
}

var _ Study = (*dxma)(nil)

type dxma struct {
	implMA
	e1, e2 Study
	idx    int
}

func (l *dxma) Update(vs ...Decimal) Decimal {
	period := l.Len()
	var e1, e2 Decimal
	for _, v := range vs {
		e1 = l.e1.Update(v)
		if l.idx < period {
			e2 = l.e2.Update(v)
			l.idx++
			continue
		}
		e2 = l.e2.Update(e1)
	}

	return e1*2 - e2
}

func (l *dxma) Len() int { return l.e1.Len() }

// TEMA - Triple Exponential Moving Average
func TEMA(period int) MovingAverage {
	return TripleMA(period, EMA)
}

// TripleMA - Triple Moving Average
func TripleMA(period int, ma MovingAverageFunc) MovingAverage {
	checkPeriod(period, 2)
	return &txma{
		e1:     ma(period),
		e2:     ma(period),
		e3:     ma(period),
		period: period,
	}
}

var _ Study = (*txma)(nil)

type txma struct {
	implMA
	e1, e2, e3 Study
	period     int
	idx        int
	max2       int
	max3       int
}

func (l *txma) Update(vs ...Decimal) Decimal {
	var e1, e2, e3 Decimal

	for _, v := range vs {
		e1 = l.e1.Update(v)
		if l.idx < l.period {
			l.idx++
			e2 = l.e2.Update(v)
			e3 = l.e3.Update(v)
			continue
		}
		e2 = l.e2.Update(e1)
		e3 = l.e3.Update(e2)
	}

	return 3*e1 - 3*e2 + e3
}

func (l *txma) Len() int { return l.period }

// TODO:
// - Trima
// - KAMA
// - MAMA/FAMA
// - T3
