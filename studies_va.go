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

// Variance - returns a multiple variable study where Update returns the variance of total values
// UpdateAll returns [variance, stddev, mean]
func Variance(period int) MultiVarStudy {
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
	mean *TA
	sum  *TA
	mode uint8
}

func (s *variance) Update(vs ...Decimal) Decimal {
	var (
		m1, m2 Decimal
		isMean = s.mode&runMean == runMean
	)
	for _, v := range vs {
		s.mean.Append(v)
		m1 = s.mean.Avg()
		if isMean {
			continue
		}

		s.sum.Append(v * v)
		m2 = s.sum.Avg()
	}

	if isMean {
		return m1
	}

	v := m2 - m1*m1
	if s.mode&runStd == runStd {
		v = v.Sqrt()
	}
	return v
}

func (s *variance) UpdateAll(vs ...Decimal) []Decimal {
	var m1, m2 Decimal
	for _, v := range vs {
		s.mean.Append(v)
		m1 = s.mean.Avg()
		s.sum.Append(v * v)
		m2 = s.sum.Avg()
	}
	v := m2 - m1*m1
	return []Decimal{v, v.Sqrt(), m1}
}

func (s *variance) Len() int               { return s.mean.Len() }
func (s *variance) LenAll() []int          { return []int{s.Len()} }
func (s *variance) ToStudy() (Study, bool) { return s, true }

func (s *variance) ToMulti() (MultiVarStudy, bool) { return s, true }
