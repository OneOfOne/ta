package ta

import (
	"math"
	"strconv"
)

type RandSource = interface {
	Int63n(int64) int64
}

func RandInRange(r RandSource, min, max F) (out F) {
	if max < min {
		min, max = max, min
	}

again:
	f := float64(r.Int63n(1<<53)) / (1 << 53)
	if f == 1 {
		goto again // resample; this branch is taken O(never)
	}
	return min + F(f)*(max-min)
}

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

func Min(vs ...F) F {
	m := F(math.MaxFloat64)
	for _, v := range vs {
		if v < m {
			m = v
		}
	}
	return m
}

func Max(vs ...F) F {
	m := -F(math.SmallestNonzeroFloat64)
	for _, v := range vs {
		if v > m {
			m = v
		}
	}
	return m
}

func MinInt(vs ...int) int {
	m := int(math.MaxInt64)
	for _, v := range vs {
		if v < m {
			m = v
		}
	}
	return m
}

func MinInt64(vs ...int64) int64 {
	m := int64(math.MaxInt64)
	for _, v := range vs {
		if v < m {
			m = v
		}
	}
	return m
}

func MinUint(vs ...uint) uint {
	m := uint(math.MaxUint64)
	for _, v := range vs {
		if v < m {
			m = v
		}
	}
	return m
}

func MinUint64(vs ...uint64) uint64 {
	m := uint64(math.MaxUint64)
	for _, v := range vs {
		if v < m {
			m = v
		}
	}
	return m
}

func MaxInt(vs ...int) int {
	m := int(math.MinInt64)
	for _, v := range vs {
		if v > m {
			m = v
		}
	}
	return m
}

func MaxInt64(vs ...int64) int64 {
	m := int64(math.MinInt64)
	for _, v := range vs {
		if v > m {
			m = v
		}
	}
	return m
}

func MaxUint(vs ...uint) uint {
	m := uint(0)
	for _, v := range vs {
		if v > m {
			m = v
		}
	}
	return m
}

func MaxUint64(vs ...uint64) uint64 {
	m := uint64(0)
	for _, v := range vs {
		if v > m {
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
