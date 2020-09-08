package ta

import (
	"math/rand"
	"sort"
	"strings"
	"testing"

	"go.oneofone.dev/ta/decimal"
)

func randSlice(size int, seed int64, min, max Decimal) *TA {
	r := rand.New(rand.NewSource(seed))
	out := NewSize(size, true)
	for i := 0; i < size; i++ {
		v := decimal.RandRange(r, min, max)
		out.Append(v)
	}

	return out
}

func TestTA(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	ta := New(make([]float64, 100), func(float64) float64 { return r.Float64() })
	sort.Float64Slice(ta.Floats()).Sort()
	if idx, _ := ta.Min(); idx != 0 {
		t.Fatalf("expected min index to be 0, got %v", idx)
	}
	if idx, _ := ta.Max(); idx != ta.Len()-1 {
		t.Fatalf("expected max index to be %v, got %v", ta.Len()-1, idx)
	}

	ta.Fill(0, 10, 42)
	for i := 0; i < 10; i++ {
		if ta.At(i) != 42 {
			t.Fatalf("expected 42, got %v", ta.At(i))
		}
	}

	if ta.At(10) == 42 {
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
		t.Fatalf("expected 3x3, 1x2 parts, got %v", parts)
	}

	t.Logf("%+v", parts)
	t.Logf("%.50f", parts)
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

func TestRing(t *testing.T) {
	t.Parallel()
	exp := New([]float64{1, 2, 3, 4, 5})
	s := NewSize(5, false).Capped()
	for i := 0; i < s.Len(); i++ {
		s.PushCapped(Decimal(i + 1))
	}
	if !s.Equal(exp) {
		t.Fatal("s != exp")
	}

	exp = New([]float64{2, 4, 6, 8, 5})
	for i := 0; i < s.Len()-1; i++ {
		s.PushCapped(Decimal((i + 1) * 2))
	}

	if !s.Equal(exp) {
		t.Fatal("s != exp")
	}
}

func TestMathFuncs(t *testing.T) {
	t.Parallel()
	type fnTest struct {
		name string
		fn   func(*TA) *TA
	}
	for _, fn := range &[...]fnTest{
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
	} {
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

	if testNothingCrossed1.Crossover(testNothingCrossed2) == true {
		t.Error("Crossover: not expected and found")
	}

	if !testCrossunder1.Crossunder(testCrossunder2) {
		t.Error("Crossunder: expected and not found")
	}

	if !testCrossover1.Crossover(testCrossover2) {
		t.Error("Crossover: expected and not found")
	}
}
