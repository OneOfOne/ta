package ta

import (
	"strconv"
)

func AbsInt(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func AbsInt64(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func MinInt(vs ...int) int {
	m := vs[0]
	for i := 1; i < len(vs); i++ {
		if v := vs[i]; v < m {
			m = v
		}
	}
	return m
}

func MinInt64(vs ...int64) int64 {
	m := vs[0]
	for i := 1; i < len(vs); i++ {
		if v := vs[i]; v < m {
			m = v
		}
	}
	return m
}

func MinUint(vs ...uint) uint {
	m := vs[0]
	for i := 1; i < len(vs); i++ {
		if v := vs[i]; v < m {
			m = v
		}
	}
	return m
}

func MinUint64(vs ...uint64) uint64 {
	m := vs[0]
	for i := 1; i < len(vs); i++ {
		if v := vs[i]; v < m {
			m = v
		}
	}
	return m
}

func MaxInt(vs ...int) int {
	m := vs[0]
	for i := 1; i < len(vs); i++ {
		if v := vs[i]; v > m {
			m = v
		}
	}
	return m
}

func MaxInt64(vs ...int64) int64 {
	m := vs[0]
	for i := 1; i < len(vs); i++ {
		if v := vs[i]; v > m {
			m = v
		}
	}
	return m
}

func MaxUint(vs ...uint) uint {
	m := vs[0]
	for i := 1; i < len(vs); i++ {
		if v := vs[i]; v > m {
			m = v
		}
	}
	return m
}

func MaxUint64(vs ...uint64) uint64 {
	m := vs[0]
	for i := 1; i < len(vs); i++ {
		if v := vs[i]; v > m {
			m = v
		}
	}
	return m
}

func CopyFloats(in []float64) []float64 {
	return append([]float64(nil), in...)
}

func checkPeriod(period, min int) {
	if period < min {
		panic("period < " + strconv.Itoa(min))
	}
}
