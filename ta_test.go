package ta

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"go.oneofone.dev/ta/decimal"
)

var (
	MaxInt = decimal.MaxInt
	MinInt = decimal.MaxInt
)

func TestMain(m *testing.M) {
	log.SetFlags(log.Lshortfile)
	pyout, _ := exec.Command("python", "-c", "import talib; print('success')").Output()
	if string(pyout[0:7]) != "success" {
		log.Println("python and talib must be installed to run most of the tests.")
		noPython = true
	}
	os.Exit(m.Run())
}

var noPython bool

func a2s(a *TA) string { // go float64 array to python list initializer string
	return strings.Replace(fmt.Sprintf("%.40f", a.v), " ", ",", -1)
}

func round(input float64) float64 {
	if input < 0 {
		return math.Ceil(input - 0.5)
	}
	return math.Floor(input + 0.5)
}

var exceptRe = regexp.MustCompile(`(?m)Exception: (.*)$`)

func check(tb testing.TB, err error, output ...[]byte) {
	if err != nil {
		tb.Helper()
		if len(output) > 0 {
			out := output[0]
			if m := exceptRe.FindAllSubmatch(out, 1); len(m) == 1 {
				tb.Fatalf("%s", m[0][1])
			} else {
				tb.Logf("%s", out)
			}
		}
		tb.Fatal(err)
	}
}

// modified version of https://github.com/markcheno/go-talib
func compare(t *testing.T, res *TA, taCall string, args ...interface{}) {
	t.Helper()
	if noPython {
		t.Skip("this test requires python")
	}
	pyprog := fmt.Sprintf(progSrc, fmt.Sprintf(taCall, args...))

	// fmt.Println(pyprog)
	cmd := exec.Command("python")
	cmd.Stdin = strings.NewReader(pyprog)
	pyOut, err := cmd.CombinedOutput()
	check(t, err, pyOut)

	tmp := strings.Fields(string(pyOut))
	pyres := NewSize(len(tmp), true)
	for _, arg := range tmp {
		if n, err := strconv.ParseFloat(arg, 64); err == nil {
			pyres.Append(Decimal(n))
		}
	}

	if res.Len() < pyres.Len() {
		pyres = pyres.Slice(-res.Len(), 0)
	}

	for i := 0; i < res.Len(); i++ {
		gr, pr := res.Get(i), pyres.Get(i)
		if gr.IsNaN() {
			gr = 0.0
		}

		if !decimal.EqualApprox(gr.Float(), pr.Float(), 1e-6) {
			t.Fatalf("[@%d] got %#v, expected %#v (diff %.20f)\ngo: %v\npy: %v",
				i, gr, pr, gr.Sub(pr).Abs(),
				res.v[MaxInt(0, i-4):MinInt(res.Len(), i+4)],
				pyres.v[MaxInt(0, i-4):MinInt(pyres.Len(), i+4)])
		}
	}
}

// Ensure that python and talib are installed and in the PATH

func testMACD(t *testing.T, fast, slow, sig int, fn MovingAverageFunc, typ string) {
	t.Run(fmt.Sprintf("%s:%v:%v:%v", typ, fast, slow, sig), func(t *testing.T) {
		out := ApplyMultiVarStudy(MACDMulti(fn(fast), fn(slow), fn(sig)), testClose)
		macd, macdsignal, macdhist := out[0], out[1], out[2]
		pyfn := fmt.Sprintf(`talib.MACDEXT(testClose, %d, talib.MA_Type.%s, %d, talib.MA_Type.%s, %d, talib.MA_Type.%s)`,
			fast, typ, slow, typ, sig, typ)
		t.Log(macd.Slice(-5, 0), macdsignal.Slice(-5, 0), macdhist.Slice(-5, 0))
		compare(t, macd, "result, macdsignal, macdhist = %s", pyfn)
		compare(t, macdsignal, "macd, result, macdhist = %s", pyfn)
		compare(t, macdhist, "macd, macdsignal, result = %s", pyfn)
	})
}

var maSteps = &[...]int{2, 3, 5, 10, 20, 32, 36, 39, 52, 71, 90, 120}

func testMA(t *testing.T, name string, fn MovingAverageFunc, maxPeriod int) {
	t.Parallel()
	for _, period := range maSteps {
		if maxPeriod > -1 && period > maxPeriod {
			t.Skipf("%s > %d overflows python", name, maxPeriod)
		}
		t.Run(strconv.Itoa(period), func(t *testing.T) {
			res, _ := testClose.MovingAverage(fn, period)
			compare(t, res, "result = talib.%s(testClose, %d)", name, period)
		})
	}
}

func testStudy(t *testing.T, name string, fn func(period int) Study, maxPeriod int) {
	t.Parallel()
	for _, period := range maSteps {
		if maxPeriod > -1 && period > maxPeriod {
			t.Skipf("%s > %d overflows python", name, maxPeriod)
		}
		t.Run(strconv.Itoa(period), func(t *testing.T) {
			res := ApplyStudy(fn(period), testClose)[0]
			compare(t, res, "result = talib.%s(testClose, %d)", name, period)
		})
	}
}

func randSlice(size int, seed int64, min, max Decimal) *TA {
	r := rand.New(rand.NewSource(seed))
	out := NewSize(size, true)
	for i := 0; i < size; i++ {
		v := decimal.RandWithSrc(r, min, max)
		out.Append(v)
	}

	return out
}
