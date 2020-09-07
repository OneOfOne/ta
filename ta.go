package ta

import (
	"unsafe"

	"go.oneofone.dev/ta/decimal"
	"gonum.org/v1/gonum/floats"
)

// Decimal is an alias to the underlying type we use.
// For now it's mostly a wrapper around float64,
// however it may change to big.Float in the future if higher accuracy is needed.
type Decimal = decimal.Decimal

const (
	Zero = Decimal(0)
	One  = Decimal(1)
)

func NewSize(size int, cap bool) *TA {
	if cap {
		return &TA{v: make([]Decimal, 0, size)}
	}
	return &TA{v: make([]Decimal, size)}
}

func New(vs []float64, transformers ...func(float64) float64) *TA {
	ds := make([]Decimal, 0, len(vs))
	for _, v := range vs {
		for _, fn := range transformers {
			v = fn(v)
		}
		ds = append(ds, Decimal(v))
	}
	return &TA{v: ds}
}

// TA the base of the techenical analysis library
type TA struct {
	v   []Decimal
	idx *int
}

func (ta *TA) index(i int) int {
	if ta.idx == nil {
		return i
	}
	return i % len(ta.v)
}

// Capped will convert TA to a buffer ring of sorts
// See `TA.Append`
func (ta *TA) Capped() *TA {
	ta.idx = new(int)
	return ta
}

// PushCapped will push v to the ringbuffer and return the previous value
// It will automatically convert the TA to Capped
func (ta *TA) PushCapped(v Decimal) (prev Decimal) {
	if ln := len(ta.v); ln < cap(ta.v) {
		if len(ta.v) > 0 {
			prev = ta.v[ln-1]
		}
		ta.v = append(ta.v, v)
		ta.idx = &ln
		return
	}

	i := 0
	if ta.idx != nil {
		i = (*ta.idx + 1) % len(ta.v)
	}
	ta.idx = &i
	prev = ta.v[i]
	ta.v[i] = v
	return
}

// Append appends v to the underlying buffer,
// if `Capped` was called it i'll act as a ring buffer rather than a slice
func (ta *TA) Append(v Decimal) *TA {
	if ta.idx == nil || len(ta.v) < cap(ta.v) {
		ta.v = append(ta.v, v)
		return ta
	}
	i := 0
	if ta.idx != nil {
		i = (*ta.idx + 1) % len(ta.v)
	}
	ta.idx = &i
	ta.v[i] = v
	return ta
}

func (ta *TA) Reverse() *TA {
	for i, j := 0, len(ta.v)-1; i < j; i, j = i+1, j-1 {
		ta.v[i], ta.v[j] = ta.v[j], ta.v[i]
	}
	return ta
}

func (ta *TA) Mapf(fn func(float64) float64, inPlace bool) *TA {
	return ta.Map(func(v Decimal) Decimal { return Decimal(fn(v.Float())) }, inPlace)
}

func (ta *TA) Map(fn func(Decimal) Decimal, inPlace bool) *TA {
	vs := ta.v
	out := ta.v[:0]

	if !inPlace {
		out = make([]Decimal, 0, ta.Len())
		ta = &TA{}
	}

	for _, v := range vs {
		out = append(out, fn(v))
	}

	ta.v = out
	return ta
}

func (ta *TA) Reduce(fn func(prev, v Decimal) Decimal, initial Decimal) Decimal {
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

func (ta *TA) Last() Decimal {
	if ln := ta.Len(); ln > 0 {
		return ta.At(ln - 1)
	}
	return 0
}

func (ta *TA) Len() int { return len(ta.v) }

// Floats returns the taice as []float64, without a copy
func (ta *TA) Floats() []float64 {
	// this must be changed if decimal != float64
	return *(*[]float64)(unsafe.Pointer(&ta.v))
}

// Data returns the underlying data slice
// if ta is capped, the data wil *not* be in order
func (ta *TA) Data() []Decimal {
	return ta.v
}

func (ta *TA) Copy() *TA {
	return &TA{v: append([]Decimal(nil), ta.v...)}
}

func (ta *TA) At(i int) Decimal {
	return ta.v[ta.index(i)]
}

func (ta *TA) SetAt(i int, v Decimal) {
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

func (ta *TA) Max() (idx int, v Decimal) {
	idx = floats.MaxIdx(ta.Floats())
	return idx, ta.v[idx]
}

func (ta *TA) Min() (idx int, v Decimal) {
	idx = floats.MinIdx(ta.Floats())
	return idx, ta.v[idx]
}

func (ta *TA) Fill(start, length int, v Decimal) *TA {
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

func (ta *TA) GroupBy(fn func(idx int, v Decimal) (group bool), aggFn func(*TA) Decimal, inPlace bool) *TA {
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
		fs = []Decimal{}
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

func (ta *TA) SplitFn(fn func(idx int, v Decimal) (split bool), copy bool) []*TA {
	var (
		ln   = len(ta.v)
		out  []*TA
		tmp  []Decimal
		last int
		vs   = ta.v
	)
	for i := 0; i < ln; i++ {
		v := vs[i]
		if fn(i, v) {
			if tmp = vs[last : i+1]; copy {
				tmp = append([]Decimal(nil), tmp...)
			}
			last = i + 1
			out = append(out, &TA{v: tmp})
		}
	}
	if last < ln {
		if tmp = vs[last:]; copy {
			tmp = append([]Decimal(nil), tmp...)
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
			tmp = append([]Decimal(nil), tmp...)
		}
		out = append(out, &TA{v: tmp})
	}

	return out
}

func (ta *TA) Sum() Decimal {
	s := floats.Sum(ta.Floats())
	return Decimal(s)
}

func (ta *TA) Avg() Decimal {
	return ta.Sum().Div(Decimal(float64(len(ta.v))))
}
