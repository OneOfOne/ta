package ta

import (
	"fmt"
	"io"
	"strings"
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

func NewCapped(size int) *TA {
	return &TA{
		v:   make([]Decimal, size),
		idx: new(int),
	}
}

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
	if i < 0 {
		i = len(ta.v) + i
	}
	if ta.idx == nil || len(ta.v) < cap(ta.v) {
		return i
	}
	idx := (*ta.idx + i) % len(ta.v)
	if idx < 0 {
		idx += len(ta.v)
	}
	return idx
}

func (ta *TA) Get(i int) Decimal {
	if i = ta.index(i + 1); i >= len(ta.v) || i < 0 {
		return 0
	}
	return ta.v[i]
}

func (ta *TA) Set(i int, v Decimal) {
	ta.v[ta.index(i)] = v
}

// Push will push v to the ringbuffer and return the previous value at the end of the buffer
func (ta *TA) Push(v Decimal) (prev Decimal) {
	if ta.idx == nil || len(ta.v) < cap(ta.v) {
		if i := len(ta.v); i > 0 {
			prev = ta.v[i-1]
		}
		ta.v = append(ta.v, v)
		if ta.idx != nil {
			*ta.idx = len(ta.v) - 1
		}
		return
	}

	i := (*ta.idx + 1) % len(ta.v)
	prev = ta.v[i]
	ta.v[i] = v
	*ta.idx = i
	return
}

// Append appends v to the underlying buffer,
// if `Capped` was called it i'll act as a ring buffer rather than a slice
func (ta *TA) Append(vs ...Decimal) *TA {
	// log.Println(ta.idx, len(ta.v)+len(vs), len(ta.v)+len(vs) <= cap(ta.v))
	if ta.idx == nil || len(ta.v)+len(vs) <= cap(ta.v) {
		ta.v = append(ta.v, vs...)
		return ta
	}

	i := *ta.idx
	for _, v := range vs {
		ta.v[i%len(ta.v)] = v
		i++
	}

	ta.idx = &i
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

	if ta.idx == nil {
		return &TA{v: ta.v[i:j:j]}
	}

	out := make([]Decimal, 0, j-i)
	idx := int(*ta.idx)
	i += idx
	j += idx
	for ; i < j; i++ {
		out = append(out, ta.Get(i))
	}

	return &TA{v: out}
}

func (ta *TA) Last() Decimal {
	if ln := ta.Len(); ln > 0 {
		return ta.Get(ln - 1)
	}
	return 0
}

func (ta *TA) Len() int {
	if ta == nil {
		return 0
	}
	return cap(ta.v)
}

// Floats returns the taice as []float64, without a copy
func (ta *TA) Floats() []float64 {
	// this must be changed if decimal != float64
	return *(*[]float64)(unsafe.Pointer(&ta.v))
}

// Raw returns the underlying data slice
// if ta is capped, the data wil *not* be in order
func (ta *TA) Raw() []Decimal {
	return ta.v
}

func (ta *TA) Copy() *TA {
	return &TA{v: append([]Decimal(nil), ta.v...)}
}

func (ta *TA) Equal(o *TA) bool {
	if ta.Len() != o.Len() {
		return false
	}

	for i := 0; i < ta.Len(); i++ {
		if ta.Get(i).NotEqual(o.Get(i)) {
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

func (ta *TA) Crossover(o *TA) bool {
	checkPeriod(ta.Len(), 3)
	if o.Len() < 3 {
		panic("o.Len() < 3")
	}

	return ta.Get(-2) <= o.Get(-2) && ta.Get(-1) > o.Get(-1)
}

func (ta *TA) Crossunder(o *TA) bool {
	checkPeriod(ta.Len(), 3)
	if o.Len() < 3 {
		panic("o.Len() < 3")
	}

	return ta.Get(-1) <= o.Get(-1) && ta.Get(-2) > o.Get(-2)
}

func (ta *TA) Fill(start, length int, v Decimal) *TA {
	if length < 1 {
		length = ta.Len() + length
	} else {
		if length = start + length; length > ta.Len() {
			length = ta.Len()
		}
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
	n := segSize - 1
	return ta.SplitFn(func(idx int, _ Decimal) bool {
		return idx%segSize == n
	}, copy)
}

func (ta *TA) Sum() Decimal {
	s := floats.Sum(ta.Floats())
	return Decimal(s)
}

func (ta *TA) Avg() Decimal {
	return ta.Sum().Div(Decimal(float64(len(ta.v))))
}

func (ta *TA) Format(f fmt.State, c rune) {
	prec, ok := f.Precision()
	if !ok {
		prec = 20
	}

	if c == 'v' {
		c = 'g'
	}

	hasPlus := f.Flag('+')
	hasZero := f.Flag('0')
	hasSpace := f.Flag(' ')
	wid, ok := f.Width()
	if !ok {
		wid = 0
	}

	if hasPlus {
		io.WriteString(f, "&TA{data: [")
	} else {
		io.WriteString(f, "[")
	}

	for i := 0; i < ta.Len(); i++ {
		if i > 0 {
			io.WriteString(f, ", ")
		}
		v := ta.Get(i).Text(byte(c), prec)
		if w := wid - len(v); w > 0 {
			if hasZero {
				io.WriteString(f, strings.Repeat("0", w))
			} else if hasSpace {
				io.WriteString(f, strings.Repeat(" ", w))
			}
		}
		io.WriteString(f, v)
	}

	if hasPlus {
		io.WriteString(f, "]")
		if ta.idx != nil {
			io.WriteString(f, ", capped: true")
		}
		io.WriteString(f, "}")
	} else {
		io.WriteString(f, "]")
	}
}
