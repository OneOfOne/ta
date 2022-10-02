package decimal

import (
	"time"
	"unsafe"

	"go.oneofone.dev/genh"
	"gonum.org/v1/gonum/floats"
)

func Abs[T genh.Signed | genh.Float](v T) T {
	if v < 0 {
		return -v
	}
	return v
}

func Min[T genh.Ordered](vs ...T) T {
	m := vs[0]
	for i := 1; i < len(vs); i++ {
		if v := vs[i]; v < m {
			m = v
		}
	}
	return m
}

func Max[T genh.Ordered](vs ...T) T {
	m := vs[0]
	for i := 1; i < len(vs); i++ {
		if v := vs[i]; v > m {
			m = v
		}
	}
	return m
}

func SliceFromFloats(in []float64, copy bool) []Decimal {
	if copy {
		in = genh.SliceClone(in)
	}
	return *(*[]Decimal)(unsafe.Pointer(&in))
}

func SliceToFloats(in []Decimal, copy bool) []float64 {
	if copy {
		in = genh.SliceClone(in)
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
