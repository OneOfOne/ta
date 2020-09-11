package ta // import "go.oneofone.dev/ta"

import (
	"fmt"
	"io"
	"strings"
	"unsafe"

	"go.oneofone.dev/ta/decimal"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
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
	idx := (*ta.idx + i + 1) % len(ta.v)
	if idx < 0 {
		idx += len(ta.v)
	}
	return idx
}

func (ta *TA) Get(i int) Decimal {
	if i = ta.index(i); i >= len(ta.v) || i < 0 {
		return 0
	}
	return ta.v[i]
}

func (ta *TA) Set(i int, v Decimal) {
	ta.v[ta.index(i)] = v
}

// Update pushes v to the end of the "buffer" and returns the previous value
// It will panic unless the ta was created with `NewCapped`
func (ta *TA) Update(v Decimal) (prev Decimal) {
	if ta.idx == nil {
		panic("requires a capped TA")
	}
	if len(ta.v) < cap(ta.v) {
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
	if ta.idx == nil || len(ta.v)+len(vs) <= cap(ta.v) {
		ta.v = append(ta.v, vs...)
		if ta.idx != nil {
			*ta.idx = len(ta.v) - 1
		}
		return ta
	}

	i := *ta.idx
	for _, v := range vs {
		i = (i + 1) % len(ta.v)
		ta.v[i] = v
	}

	*ta.idx = i
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
	} else if j < 0 {
		j = MinInt(ln, i-j)
	}

	if ta.idx == nil {
		return &TA{v: ta.v[i:j:j]}
	}

	out := make([]Decimal, 0, j-i)
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

// Floats returns the taice as []float64, without a copy if not capped
// if ta is capped, it'll create a copy
func (ta *TA) Floats() []float64 {
	if ta.idx == nil {
		// this must be changed if decimal != float64
		return *(*[]float64)(unsafe.Pointer(&ta.v))
	}

	out := make([]float64, 0, ta.Len())
	for i := 0; i < ta.Len(); i++ {
		out = append(out, ta.Get(i).Float())
	}
	return out
}

func (ta *TA) floats() []float64 {
	return *(*[]float64)(unsafe.Pointer(&ta.v))
}

func (ta *TA) Uncapped() *TA {
	if ta.idx == nil {
		return ta.Copy()
	}
	fs := ta.Floats()
	return &TA{v: *(*[]Decimal)(unsafe.Pointer(&fs))}
}

// Raw returns the underlying data slice
// if ta is capped, the data will *not* be in order
func (ta *TA) Raw() []Decimal {
	return ta.v
}

func (ta *TA) Copy() *TA {
	return &TA{v: append([]Decimal(nil), ta.v...), idx: ta.idx}
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
		out  *TA
	)

	if !inPlace {
		fs = []Decimal{}
	}

	for i, last := 0, 0; i < ln; i++ {
		v := ta.Get(i)
		if fn(i, v) {
			out = ta.Slice(last, i+1)
			fs = append(fs, aggFn(out))
			last = i + 1
		}
	}

	if last < ln {
		out = ta.Slice(last, 0)
		fs = append(fs, aggFn(out))
	}

	out.v = fs[:len(fs):len(fs)]
	return out
}

func (ta *TA) SplitFn(fn func(idx int, v Decimal) (split bool), copy bool) []*TA {
	var (
		ln   = len(ta.v)
		out  []*TA
		last int
	)

	for i := 0; i < ln; i++ {
		v := ta.Get(i)
		if fn(i, v) {
			tmp := ta.Slice(last, i+1)
			if copy && ta.idx == nil {
				tmp = tmp.Copy()
			}
			last = i + 1
			out = append(out, tmp)
		}
	}

	if last < ln {
		tmp := ta.Slice(last, 0)
		if copy && ta.idx == nil {
			tmp = tmp.Copy()
		}
		out = append(out, tmp)
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
	s := floats.Sum(ta.floats())
	return Decimal(s)
}

// CumSum finds the cumulative sum of the ta
func (ta *TA) CumSum() *TA {
	dst := floats.CumSum(make([]float64, ta.Len()), ta.Floats())
	return &TA{v: *(*[]Decimal)(unsafe.Pointer(&dst))}
}

// CumProd finds the cumulative product of the ta
func (ta *TA) CumProd() *TA {
	dst := floats.CumProd(make([]float64, ta.Len()), ta.Floats())
	return &TA{v: *(*[]Decimal)(unsafe.Pointer(&dst))}
}

func (ta *TA) Product() Decimal {
	return Decimal(floats.Prod(ta.Floats()))
}

func (ta *TA) Avg() Decimal {
	return ta.Sum() / Decimal(len(ta.v))
}

func (ta *TA) Dot(o *TA) Decimal {
	return Decimal(floats.Dot(ta.Floats(), o.Floats()))
}

func (ta *TA) StdDevSum() Decimal {
	return Decimal(stat.StdDev(ta.Floats(), nil))
}

func (ta *TA) VarianceSum() Decimal {
	return Decimal(stat.Variance(ta.Floats(), nil))
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
