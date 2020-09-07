package ta

import (
	"math/rand"
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
