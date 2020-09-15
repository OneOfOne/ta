package ta

import (
	"math/rand"
	"sort"
	"strings"
	"testing"

	"go.oneofone.dev/ta/decimal"
)

func TestTA(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	ta := New(make([]float64, 100), func(float64) float64 { return r.Float64() })
	sort.Float64Slice(ta.Floats()).Sort()
	if idx := ta.MinIndex(); idx != 0 {
		t.Fatalf("expected min index to be 0, got %v", idx)
	}
	if idx := ta.MaxIndex(); idx != ta.Len()-1 {
		t.Fatalf("expected max index to be %v, got %v", ta.Len()-1, idx)
	}

	ta.Fill(0, 10, 42)
	for i := 0; i < 10; i++ {
		if ta.Get(i) != 42 {
			t.Fatalf("expected 42, got %v", ta.Get(i))
		}
	}

	if ta.Get(10) == 42 {
		t.Fatal("expected a random number, got 42")
	}

	if n := ta.Slice(0, 10).Sum(); n != 420 {
		t.Fatalf("expected to blaze it, got %v", n)
	}

	if n := ta.Slice(0, 10).Avg(); n != 42 {
		t.Fatalf("expected 42, got %v", n)
	}

	parts := ta.Slice(0, 11).Split(3, false)
	if len(parts) != 4 {
		t.Fatalf("expected 4 parts, got %v %v", len(parts), parts)
	}

	if parts[0].Len() != 3 || parts[1].Len() != 3 || parts[2].Len() != 3 || parts[3].Len() != 2 {
		t.Fatalf("expected 3x3, 1x2 parts, got %v %v", len(parts), parts)
	}
}

func TestGroupBy(t *testing.T) {
	const N = 60 * 16 // assuming 4am -> 8pm trading day
	s := randSlice(N, 42, 111, 120)
	expected := New([]float64{
		114.8577953397332, 115.32734645207431, 115.02437162968172, 115.45831810458712, 115.33003237997008, 116.06364289628996, 115.21384897360117,
		115.68005825728478, 115.13575339798609, 115.5782899338858, 115.80296018224767, 115.88345372453232, 115.24272441950352, 115.13836671287244,
		115.79672239067357, 115.44007649128464,
	})
	agg := s.GroupBy(func(idx int, _ Decimal) bool { // convert minute chart to hour
		return idx > 0 && idx%60 == 0
	}, nil, false)

	if !agg.Equal(expected) {
		t.Fatal("not equal")
	}
}

func TestCapped(t *testing.T) {
	t.Parallel()
	exp := New([]float64{496, 497, 498, 499, 500})
	s := NewCapped(5)
	for i := 0; i < s.Len()*100; i++ {
		// m1 := int(s.Push(Decimal(i + 1)))
		s.Update(Decimal((i + 1)))
		// t.Log(s.v)
		if i < 10 {
			continue
		}
		for j := 0; j < s.Len(); j++ {
			x := (s.Len() - j) - 2
			exp := i - x
			if v := int(s.Get(j)); v != exp {
				t.Fatal(i, j, x, v, i/5, exp, s.v)
			}
		}
	}
	if !s.Equal(exp) {
		t.Fatal("s != exp", s, s.v, exp, exp.v)
	}

	exp = New([]float64{2, 4, 6, 8, 10})
	for i := 0; i < s.Len(); i++ {
		s.Update(Decimal((i + 1) * 2))
	}

	if !s.Equal(exp) {
		t.Fatal("s != exp", s.v, exp, s.Get(0))
	}

	ta := NewCapped(3)
	for i := 0; i < 10; i++ {
		t.Log(i, ta.v)
		ta.Update(Decimal(i + 1))
	}
	t.Log(ta.v)
	if v := ta.Raw(); !decimal.SliceEqual(v, []Decimal{9, 10, 8}) {
		t.Fatalf("wrong slice value x1: %+v", v)
	}

	if v := ta.Slice(0, 3).Raw(); !decimal.SliceEqual(v, []Decimal{8, 9, 10}) {
		t.Fatalf("wrong slice value x2: %+v", v)
	}
	v := ta.Slice(0, 3)
	t.Log(ta.v, v.v, ta.idx, ta, v)
}

func TestMathFuncs(t *testing.T) {
	t.Parallel()
	tests := &[...]struct {
		name string
		fn   func(*TA) *TA
	}{
		{"Acos", (*TA).Acos},
		{"Asin", (*TA).Asin},
		{"Atan", (*TA).Atan},
		{"Ceil", (*TA).Ceil},
		{"Cos", (*TA).Cos},
		{"Cosh", (*TA).Cosh},
		{"Exp", (*TA).Exp},
		{"Floor", (*TA).Floor},
		{"Ln", (*TA).Ln},
		{"Log10", (*TA).Log10},
		{"Sin", (*TA).Sin},
		{"Sinh", (*TA).Sinh},
		{"Sqrt", (*TA).Sqrt},
		{"Tan", (*TA).Tan},
		{"Tanh", (*TA).Tanh},
	}

	for _, fn := range tests {
		t.Run(fn.name, func(t *testing.T) {
			res := fn.fn(testRand)
			compare(t, res, "result = talib.%s(testRand)", strings.ToUpper(fn.name))
		})
	}
}

func TestMathOps(t *testing.T) {
	t.Parallel()
	type fnTest struct {
		name string
		fn   func(*TA, *TA) *TA
	}
	for _, fn := range &[...]fnTest{
		{"Add", (*TA).Add},
		{"Sub", (*TA).Sub},
		{"Mult", (*TA).Mul},
		{"Div", (*TA).Div},
	} {
		t.Run(fn.name, func(t *testing.T) {
			res := fn.fn(testHigh, testLow)
			compare(t, res, "result = talib.%s(testHigh, testLow)", strings.ToUpper(fn.name))
		})
	}
}

func TestCrossOverUnder(t *testing.T) {
	t.Parallel()
	if testCrossunder1.Crossover(testCrossunder2) {
		t.Error("Crossover: not expected and found")
	}

	if testNothingCrossed1.Crossover(testNothingCrossed2) {
		t.Error("Crossover: not expected and found")
	}

	if !testCrossunder1.Crossunder(testCrossunder2) {
		t.Error("Crossunder: expected and not found")
	}

	if !testCrossover1.Crossover(testCrossover2) {
		t.Error("Crossover: expected and not found")
	}
}
