package ta

// MinMax returns the min/max over a period
// Update returns min
// UpdateAll returns [min, max]
func MinMax(period int) MultiVarStudy {
	return &minmax{data: NewCapped(period)}
}

func Min(period int) Study {
	return &minmax{data: NewCapped(period)}
}

func Max(period int) Study {
	return &minmax{data: NewCapped(period), isMax: true}
}

type minmax struct {
	data  *TA
	isMax bool
}

func (s *minmax) Update(vs ...Decimal) Decimal {
	ta := s.data.Append(vs...)
	if s.isMax {
		return ta.Max()
	}
	return ta.Min()
}

func (s *minmax) UpdateAll(vs ...Decimal) []Decimal {
	ta := s.data.Append(vs...)
	return []Decimal{ta.Min(), ta.Max()}
}

func (s *minmax) Len() int { return s.data.Len() }
func (s *minmax) LenAll() []int {
	ln := s.Len()
	return []int{ln, ln}
}

func (s *minmax) ToStudy() (Study, bool) { return s, true }

func (s *minmax) ToMulti() (MultiVarStudy, bool) { return s, true }

// BBands alias for BollingerBands(period, d, -d, nil)
func BBands(period int) MultiVarStudy {
	d := Decimal(period)
	return BollingerBands(period, d, d, nil)
}

// BBandsLimits alias for BollingerBands(period, up, down, nil)
func BBandsLimits(period int, up, down Decimal) MultiVarStudy {
	return BollingerBands(period, up, down, nil)
}

// BollingerBands returns a Bollinger Bands study, if ma is nil, the mid will be the mean
// Update will return the upper bound
// UpdateAll returns [upper, mid, lower]
func BollingerBands(period int, up, down Decimal, ma MovingAverageFunc) MultiVarStudy {
	if down > 0 {
		down = -down
	}
	bb := &bbands{
		std:  newVar(period, runStd),
		up:   up,
		down: down,
	}

	if ma != nil {
		bb.ext = ma(period)
	}

	return bb
}

type bbands struct {
	ext  MovingAverage
	std  *variance
	up   Decimal
	down Decimal
}

func (s *bbands) Update(vs ...Decimal) Decimal {
	return s.UpdateAll(vs...)[0]
}

func (s *bbands) UpdateAll(vs ...Decimal) []Decimal {
	var sd, base Decimal
	if s.ext == nil {
		data := s.std.UpdateAll(vs...)
		sd, base = data[1], data[2]
	} else {
		sd = s.std.Update(vs...)
		base = s.ext.Update(vs...)
	}
	return []Decimal{base + sd*s.up, base, base + sd*s.down}
}

func (s *bbands) Len() int { return s.std.Len() }
func (s *bbands) LenAll() []int {
	ln := s.Len()
	return []int{ln, ln, ln}
}
func (s *bbands) ToStudy() (Study, bool) { return s, true }

func (s *bbands) ToMulti() (MultiVarStudy, bool) { return s, true }
