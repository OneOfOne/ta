package ta

import (
	"unsafe"

	"go.oneofone.dev/ta/decimal"
	"gonum.org/v1/gonum/floats"
)

type Live interface {
	Setup(d TA) TA
	Update(v F) F
	Len() int
	Clone() Live
}

// F is an alias to the underlying type we use.
// For now it's mostly a wrapper around float64,
// however it may change to big.Float in the future if higher accuracy is needed.
type F = decimal.Decimal

const (
	Zero = F(0)
	One  = F(1)
)

// TA the base of the techenical analysis library
type TA []F

func (sl *TA) AppendToRing(vs ...F) TA {
	s := *sl
	s = append(s[:0], s[len(vs):]...)
	s = append(s, vs...)
	*sl = s
	return s
}

func (sl TA) Reverse() TA {
	for i, j := 0, len(sl)-1; i < j; i, j = i+1, j-1 {
		sl[i], sl[j] = sl[j], sl[i]
	}
	return sl
}

func (sl TA) Apply(fn func(F) F, inPlace bool) TA {
	out := sl[:0]

	if !inPlace {
		out = make(TA, 0, len(sl))
	}

	for _, v := range sl {
		out = append(out, fn(v))
	}

	return out
}

func (sl TA) Slice(i, j int) TA {
	if i < 0 {
		i = len(sl) + i
	}
	if j == 0 {
		j = len(sl)
	} else {
		j = MinInt(len(sl), AbsInt(i+j))
	}
	return sl[i:j:j]
}

func (sl TA) Last() F {
	if len(sl) == 0 {
		return 0
	}
	return sl[len(sl)-1]
}

func (sl TA) Len() int { return len(sl) }

// Floats returns the slice as []float64, without a copy
func (sl TA) Floats() []float64 {
	// this must be changed if decimal != float64
	return *(*[]float64)(unsafe.Pointer(&sl))
}

func (sl TA) Copy() TA {
	return append(TA(nil), sl...)
}

func (sl TA) Equal(o TA) bool {
	if len(sl) != len(o) {
		return false
	}

	for i := range sl {
		if sl[i].NotEqual(o[i]) {
			return false
		}
	}

	return true
}

func (sl TA) Add(o TA) TA {
	cp := sl.Copy()
	floats.Add(cp.Floats(), o.Floats())
	return cp
}

func (sl TA) Sub(o TA) TA {
	cp := sl.Copy()
	floats.Sub(cp.Floats(), o.Floats())
	return cp
}

func (sl TA) Mul(o TA) TA {
	cp := sl.Copy()
	floats.Mul(cp.Floats(), o.Floats())
	return cp
}

func (sl TA) Div(o TA) TA {
	cp := sl.Copy()
	floats.Div(cp.Floats(), o.Floats())
	return cp
}

func (sl TA) Max() (idx int, v F) {
	idx = floats.MaxIdx(sl.Floats())
	return idx, sl[idx]
}

func (sl TA) Min() (idx int, v F) {
	idx = floats.MinIdx(sl.Floats())
	return idx, sl[idx]
}

func (sl TA) Fill(start, length int, v F) TA {
	if length < 1 {
		length = len(sl) + length
	} else {
		length = start + length
	}
	for i := start; i < length; i++ {
		sl[i] = v
	}
	return sl
}

func (sl TA) Split(segSize int, copy bool) []TA {
	if segSize < 1 {
		return nil
	}
	if segSize >= len(sl) {
		return []TA{sl}
	}

	n := len(sl) / segSize
	if len(sl)%segSize != 0 {
		n++
	}
	out := make([]TA, 0, n)

	for i, j := 0, segSize; i < len(sl); i, j = j, j+segSize {
		if j > len(sl) {
			j = len(sl)
		}
		if copy {
			out = append(out, sl[i:j].Copy())
		} else {
			out = append(out, sl[i:j])
		}
	}

	return out
}

func (sl TA) Sum() F {
	s := floats.Sum(sl.Floats())
	return F(s)
}

func (sl TA) Avg() F {
	return sl.Sum().Div(F(float64(len(sl))))
}
