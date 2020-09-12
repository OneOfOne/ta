package ta

const (
	// I promise this will make sense one day
	runMean uint8 = 1 << iota
	runVar
	runStd
)

// Mean - returns an updatable study where Update returns the mean of total values
func Mean(period int) Study {
	checkPeriod(period, 2)
	return newVar(period, runMean)
}

// StdDev - returns an updatable study where Update returns the standard deviation of total values
func StdDev(period int) Study {
	checkPeriod(period, 2)
	return newVar(period, runStd)
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

var _ Study = (*variance)(nil)

type variance struct {
	noMulti
	mean *TA
	sum  *TA
	mode uint8
}

func (l *variance) Update(vs ...Decimal) Decimal {
	var (
		m1, m2 Decimal
		isMean = l.mode&runMean == runMean
	)
	for _, v := range vs {
		l.mean.Append(v)
		m1 = l.mean.Avg()
		if isMean {
			continue
		}

		l.sum.Append(v * v)
		m2 = l.sum.Avg()
	}

	if isMean {
		return m1
	}

	v := m2 - m1*m1
	if l.mode&runStd == runStd {
		v = v.Sqrt()
	}
	return v
}

func (l *variance) Len() int {
	return l.mean.Len()
}
