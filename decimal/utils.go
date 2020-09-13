package decimal

import "unsafe"

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
	return append(make([]float64, 0, len(in)), in...)
}

func CopySlice(in []Decimal) []Decimal {
	return append(make([]Decimal, 0, len(in)), in...)
}

func SliceFromFloats(in []float64, copy bool) []Decimal {
	if copy {
		in = CopyFloats(in)
	}
	return *(*[]Decimal)(unsafe.Pointer(&in))
}

func SliceToFloats(in []Decimal, copy bool) []float64 {
	if copy {
		in = CopySlice(in)
	}
	return *(*[]float64)(unsafe.Pointer(&in))
}
