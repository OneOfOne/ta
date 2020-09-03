package ta

import "strconv"

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

func Min(a, b F) F {
	if a < b {
		return a
	}
	return b
}

func Max(a, b F) F {
	if a > b {
		return a
	}
	return b
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func MinUint(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}

func MinUint64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func MaxUint(a, b uint) uint {
	if a > b {
		return a
	}
	return b
}

func MaxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func CopyFloats(in []float64) []float64 {
	return append([]float64(nil), in...)
}

func checkPeriod(period, min int) {
	if period < min {
		panic("period < " + strconv.Itoa(min))
	}
}
