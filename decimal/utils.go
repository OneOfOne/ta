package decimal

import (
	"time"
	"unsafe"

	"gonum.org/v1/gonum/floats"
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

func AggPipe(aggPeriod time.Duration, in <-chan Decimal) <-chan Decimal {
	ch := make(chan Decimal, 100)
	buf := make([]float64, 0, 600)
	go func() {
		t := time.Now()
		for v := range in {
			if n := time.Now(); n.Sub(t) >= aggPeriod {
				avg := floats.Sum(buf) / float64(len(buf))
				ch <- Decimal(avg)
				buf, t = buf[:0], n
			}
			buf = append(buf, v.Float())
		}
		if len(buf) > 0 {
			avg := floats.Sum(buf) / float64(len(buf))
			ch <- Decimal(avg)
		}
		close(ch)
	}()
	return ch
}
