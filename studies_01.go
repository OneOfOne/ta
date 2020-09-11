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

/*
1 197.63999999999998636 201.28000000000000114 199.45999999999997954 39787.60399999999936 3.3124000000098021701
2 195.78000000000000114 197.63999999999998636 196.70999999999997954 38695.688999999998487 0.86490000000776490197
3 198.21999999999999886 195.78000000000000114 197 38810.488400000002002 1.4884000000020023435
4 201.74000000000000909 198.21999999999999886 199.98000000000001819 39995.097999999998137 3.0975999999936902896
5 200.12000000000000455 201.74000000000000909 200.93000000000000682 40373.520999999993364 0.6560999999928753823
6 198.55000000000001137 200.12000000000000455 199.33500000000000796 39735.05844999999681 0.61622499999066349119
7 197.99000000000000909 198.55000000000001137 198.27000000000001023 39311.071299999995972 0.078399999991233926266
8 196.80000000000001137 197.99000000000000909 197.39500000000001023 38965.140049999994517 0.3540249999932711944
*/
func (l *variance) xUpdate(v Decimal) Decimal {
	// if l.idx += l.period; l.idx == l.period {
	// 	l.meanx = v
	// 	return v
	// }

	// n := l.idx

	// delta := v - l.meanx
	// deltaN := delta / n
	// l.meanx += deltaN

	// if l.mode == 0 {
	// 	return l.meanx
	// }

	// if l.period > 1 {
	// 	delta = delta * (1 / l.period)
	// }
	// sum := delta * deltaN * (n - 1)
	// psum := l.sumx

	// l.sumx += sum

	// if l.mode&runUnbiased == runUnbiased {
	// 	v = l.sumx / (n - 1)
	// } else {
	// 	v = l.sumx / n
	// }

	// log.Println(v)
	// if l.mode&runVar == runVar {
	// 	return v
	// }

	// if l.mode&runSqrt == runSqrt {
	// 	return v.Sqrt()
	// }

	// deltaN2 := deltaN * deltaN
	// l.kurt += sum*deltaN2*(n*n-3*n+3) + 6*deltaN2*psum - 4*deltaN*l.skew
	// l.skew += sum*deltaN*(n-2) - 3*deltaN*psum

	// if l.mode&runSkew == runSkew {
	// 	// sqrt(double(n)) * M3/ pow(M2, 1.5);
	// 	return l.idx.Sqrt() * l.skew / l.sumx.Pow(1.5)
	// }

	// if l.mode&runKurt == runKurt {
	// 	//  double(n)*M4 / (M2*M2) - 3.0;
	// 	return n*l.kurt/(l.sumx*l.sumx) - 3
	// }

	// panic("invalid mode")
	return 0
}

/*
    def push(self, x):
        self.n += 1

        if self.n == 1:
            self.old_m = self.new_m = x
            self.old_s = 0
        else:
            self.new_m = self.old_m + (x - self.old_m) / self.n
            self.new_s = self.old_s + (x - self.old_m) * (x - self.new_m)

            self.old_m = self.new_m
            self.old_s = self.new_s

    def mean(self):
        return self.new_m if self.n else 0.0

    def variance(self):
        return self.new_s / (self.n - 1) if self.n > 1 else 0.0

    def standard_deviation(self):
		return math.sqrt(self.variance())
*/
func (l *variance) Len() int {
	return l.mean.Len()
}
