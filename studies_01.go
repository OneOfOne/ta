package ta

const (
	runMean uint8 = 1 << iota
	runVar
	runStd
	runSkew
	runKurt
)

// Mean - returns an updatable study where Update returns the mean of total values
func (ta *TA) Mean(period int) (*TA, Study) {
	s := Mean(period)
	return s.Setup(ta), s
}

// Mean - returns an updatable study where Update returns the mean of total values
func Mean(period int) Study {
	checkPeriod(period, 2)
	return newVar(period, runMean)
}

// StdDev - returns an updatable study where Update returns the standard deviation of total values
func (ta *TA) StdDev(period int) (*TA, Study) {
	s := StdDev(period)
	return s.Setup(ta), s
}

// StdDev - returns an updatable study where Update returns the standard deviation of total values
func StdDev(period int) Study {
	checkPeriod(period, 2)
	return newVar(period, runStd)
}

// Variance - returns an updatable study where Update returns the variance of total values
func (ta *TA) Variance(period int) (*TA, Study) {
	s := Variance(period)
	return s.Setup(ta), s
}

// Variance - returns an updatable study where Update returns the variance of total values
func Variance(period int) Study {
	checkPeriod(period, 2)
	return newVar(period, runVar)
}

func newVar(period int, mode uint8) *variance {
	v := &variance{
		mean: NewCapped(period),
		mode: mode,
	}
	if mode&runMean != runMean {
		v.sum = NewCapped(period)
	}
	return v
}

type variance struct {
	mean *TA
	sum  *TA
	mode uint8
}

func (l *variance) Setup(d *TA) *TA {
	d = d.Map(l.Update, false)
	return d.Slice(-l.mean.Len(), 0)
}

func (l *variance) Update(v Decimal) Decimal {
	l.mean.Append(v)
	m1 := l.mean.Avg()
	if l.mode&runMean == runMean {
		return m1
	}

	l.sum.Append(v * v)
	m2 := l.sum.Avg()

	if v = m2 - m1*m1; l.mode&runStd == runStd {
		v = v.Sqrt()
	}
	return v
}

func (l *variance) Len() int {
	return l.mean.Len()
}
