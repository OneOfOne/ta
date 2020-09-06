package ta

import (
	"unsafe"

	"go.oneofone.dev/ta/decimal"
	"gonum.org/v1/gonum/floats"
)

type Live interface {
	Setup(d *TA) *TA
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

func NewSize(size int, cap bool) *TA {
	if cap {
		return &TA{v: make([]F, 0, size), idx: -1}
	}
	return &TA{v: make([]F, size), idx: -1}
}

func New(vs []float64, transformers ...func(float64) float64) *TA {
	ds := make([]F, 0, len(vs))
	for _, v := range vs {
		for _, fn := range transformers {
			v = fn(v)
		}
		ds = append(ds, F(v))
	}
	return &TA{v: ds, idx: -1}
}

// TA the base of the techenical analysis library
type TA struct {
	v   []F
	idx int
}

func (ta *TA) index(i int) int {
	if ta.idx == -1 {
		return i
	}
	return i % len(ta.v)
}

func (ta *TA) Append(v F, rotate bool) *TA {
	if !rotate {
		ta.v = append(ta.v, v)
		return ta
	}
	if ta.idx == -1 {
		ta.idx = 0
	} else {
		ta.idx = (ta.idx + 1) % len(ta.v)
	}
	ta.v[ta.idx] = v
	return ta
}

func (ta *TA) Reverse() *TA {
	for i, j := 0, len(ta.v)-1; i < j; i, j = i+1, j-1 {
		ta.v[i], ta.v[j] = ta.v[j], ta.v[i]
	}
	return ta
}

func (ta *TA) Mapf(fn func(float64) float64, inPlace bool) *TA {
	return ta.Map(func(v F) F { return F(fn(v.Float())) }, inPlace)
}

func (ta *TA) Map(fn func(F) F, inPlace bool) *TA {
	vs := ta.v
	out := ta.v[:0]

	if !inPlace {
		out = make([]F, 0, ta.Len())
		ta = &TA{}
	}

	for _, v := range vs {
		out = append(out, fn(v))
	}

	ta.v = out
	return ta
}

func (ta *TA) Reduce(fn func(prev, v F) F, initial F) F {
	for _, v := range ta.v {
		initial = fn(initial, v)
	}
	return initial
}

func (ta *TA) Slice(i, j int) *TA {
	ln := ta.Len()
	if i < 0 {
		i = ln + i
	}
	if j == 0 {
		j = ln
	} else {
		j = MinInt(ln, AbsInt(i+j))
	}
	return &TA{v: ta.v[i:j:j]}
}

func (ta *TA) Last() F {
	if ln := ta.Len(); ln > 0 {
		return ta.v[ln-1]
	}
	return 0
}

func (ta *TA) Len() int { return len(ta.v) }

// Floats returns the taice as []float64, without a copy
func (ta *TA) Floats() []float64 {
	// this must be changed if decimal != float64
	return *(*[]float64)(unsafe.Pointer(&ta.v))
}

func (ta *TA) Copy() *TA {
	return &TA{v: append([]F(nil), ta.v...), idx: -1}
}

func (ta *TA) At(i int) F {
	return ta.v[ta.index(i)]
}

func (ta *TA) SetAt(i int, v F) {
	ta.v[ta.index(i)] = v
}

func (ta *TA) Equal(o *TA) bool {
	if ta.Len() != o.Len() {
		return false
	}

	for i := 0; i < ta.Len(); i++ {
		if ta.At(i).NotEqual(o.At(i)) {
			return false
		}
	}

	return true
}

func (ta *TA) Add(o *TA) *TA {
	cp := ta.Copy()
	floats.Add(cp.Floats(), o.Floats())
	return cp
}

func (ta *TA) Sub(o *TA) *TA {
	cp := ta.Copy()
	floats.Sub(cp.Floats(), o.Floats())
	return cp
}

func (ta *TA) Mul(o *TA) *TA {
	cp := ta.Copy()
	floats.Mul(cp.Floats(), o.Floats())
	return cp
}

func (ta *TA) Div(o *TA) *TA {
	cp := ta.Copy()
	floats.Div(cp.Floats(), o.Floats())
	return cp
}

func (ta *TA) Max() (idx int, v F) {
	idx = floats.MaxIdx(ta.Floats())
	return idx, ta.v[idx]
}

func (ta *TA) Min() (idx int, v F) {
	idx = floats.MinIdx(ta.Floats())
	return idx, ta.v[idx]
}

func (ta *TA) Fill(start, length int, v F) *TA {
	if length < 1 {
		length = ta.Len() + length
	} else {
		length = start + length
	}

	for i := start; i < length; i++ {
		ta.v[ta.index(i)] = v
	}
	return ta
}

func (ta *TA) GroupBy(fn func(idx int, v F) (group bool), aggFn func(*TA) F, inPlace bool) *TA {
	if aggFn == nil {
		aggFn = (*TA).Avg
	}

	var (
		fs   = ta.v[:0]
		last int
		ln   = ta.Len()
		vs   = ta.v
		out  = ta
	)

	if !inPlace {
		fs = []F{}
		out = &TA{}
	}

	for i, last := 0, 0; i < ln; i++ {
		v := vs[i]
		if fn(i, v) {
			out.v = vs[last : i+1]
			fs = append(fs, aggFn(out))
			last = i + 1
		}
	}

	if last < ln {
		out.v = vs[last:]
		fs = append(fs, aggFn(out))
	}

	out.v = fs[:len(fs):len(fs)]
	return out
}

func (ta *TA) SplitFn(fn func(idx int, v F) (split bool), copy bool) []*TA {
	var (
		ln   = len(ta.v)
		out  []*TA
		tmp  []F
		last int
		vs   = ta.v
	)
	for i := 0; i < ln; i++ {
		v := vs[i]
		if fn(i, v) {
			if tmp = vs[last : i+1]; copy {
				tmp = append([]F(nil), tmp...)
			}
			last = i + 1
			out = append(out, &TA{v: tmp})
		}
	}
	if last < ln {
		if tmp = vs[last:]; copy {
			tmp = append([]F(nil), tmp...)
		}
		out = append(out, &TA{v: tmp})
	}
	return out[:len(out):len(out)]
}

func (ta *TA) Split(segSize int, copy bool) []*TA {
	if segSize < 1 {
		return nil
	}
	ln := len(ta.v)
	if segSize >= ln {
		return []*TA{ta}
	}

	n := ln / segSize
	if ln%segSize != 0 {
		n++
	}
	out := make([]*TA, 0, n)

	for i, j := 0, segSize; i < ln; i, j = j, j+segSize {
		if j > ln {
			j = ln
		}
		tmp := ta.v[i:j]
		if copy {
			tmp = append([]F(nil), tmp...)
		}
		out = append(out, &TA{v: tmp})
	}

	return out
}

func (ta *TA) Sum() F {
	s := floats.Sum(ta.Floats())
	return F(s)
}

func (ta *TA) Avg() F {
	return ta.Sum().Div(F(float64(len(ta.v))))
}
