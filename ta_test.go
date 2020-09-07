package ta

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"go.oneofone.dev/ta/decimal"
)

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
		gr, pr := res.At(i), pyres.At(i)
		if gr.IsNaN() {
			gr = 0.0
		}

		if !decimal.EqualApprox(gr.Float(), pr.Float(), 1e-6) {
			t.Fatalf("[@%d] got %#v, expected %#v (diff %.20f)\ngo: %v\npy: %v",
				i, gr, pr, gr.Sub(pr).Abs(),
				res.Slice(MaxInt(0, i-4), MinInt(res.Len(), i+4)),
				pyres.Slice(MaxInt(0, i-4), MinInt(pyres.Len(), i+4)))
		}
	}
}

// Ensure that python and talib are installed and in the PATH
func TestMain(m *testing.M) {
	log.SetFlags(log.Lshortfile)
	pyout, _ := exec.Command("python", "-c", "import talib; print('success')").Output()
	if string(pyout[0:7]) != "success" {
		log.Println("python and talib must be installed to run most of the tests.")
		noPython = true
	}
	os.Exit(m.Run())
}

func TestSMA(t *testing.T) {
	testMA(t, "SMA", SMA, 120)
}

func BenchmarkSMA(b *testing.B) {
	benchMA(b, "SMA", 10, SMA)
}

func TestEMA(t *testing.T) {
	testMA(t, "EMA", EMA, -1)
}

func BenchmarkEMA(b *testing.B) {
	benchMA(b, "EMA", 10, EMA)
}

func TestWMA(t *testing.T) {
	testMA(t, "WMA", WMA, -1)
}

func BenchmarkWMA(b *testing.B) {
	benchMA(b, "WMA", 10, WMA)
}

func TestDEMA(t *testing.T) {
	testMA(t, "DEMA", DEMA, 36)
}

func BenchmarkDEMA(b *testing.B) {
	benchMA(b, "DEMA", 10, DEMA)
}

func TestTEMA(t *testing.T) {
	testMA(t, "TEMA", TEMA, 32)
}

func BenchmarkTEMA(b *testing.B) {
	benchMA(b, "TEMA", 10, TEMA)
}

func TestRSI(t *testing.T) {
	testMA(t, "RSI", RSI, 120)
}

func BenchmarkRSI(b *testing.B) {
	benchMA(b, "RSI", 10, RSI)
}

func TestMACD(t *testing.T) {
	t.Parallel()
	tests := &[...]*[3]int{{12, 26, 9}, {6, 12, 6}, {10, 17, 12}}
	for _, ts := range tests {
		testMACD(t, ts[0], ts[1], ts[2], SMA, "SMA")
		testMACD(t, ts[0], ts[1], ts[2], EMA, "EMA")
		testMACD(t, ts[0], ts[1], ts[2], WMA, "WMA")
		testMACD(t, ts[0], ts[1], ts[2], DEMA, "DEMA")
		testMACD(t, ts[0], ts[1], ts[2], TEMA, "TEMA")
	}
}

func testMACD(t *testing.T, fast, slow, sig int, fn MovingAverageFunc, typ string) {
	t.Run(fmt.Sprintf("%s:%v:%v:%v", typ, fast, slow, sig), func(t *testing.T) {
		macd, macdsignal, macdhist, _ := testClose.MACDExt(fn(fast), fn(slow), fn(sig))
		pyfn := fmt.Sprintf(`talib.MACDEXT(testClose, %d, talib.MA_Type.%s, %d, talib.MA_Type.%s, %d, talib.MA_Type.%s)`,
			fast, typ, slow, typ, sig, typ)
		compare(t, macd, "result, macdsignal, macdhist = %s", pyfn)
		compare(t, macdsignal, "macd, result, macdhist = %s", pyfn)
		compare(t, macdhist, "macd, macdsignal, result = %s", pyfn)
	})
}

var maSteps = &[...]int{2, 3, 5, 10, 20, 32, 36, 39, 52, 71, 90, 120, 132, 180}

func testMA(t *testing.T, name string, fn MovingAverageFunc, maxPeriod int) {
	t.Helper()
	t.Parallel()
	for _, period := range maSteps {
		if maxPeriod > -1 && period > maxPeriod {
			t.Skipf("%s > %d overflows python", name, maxPeriod)
		}
		t.Run(strconv.Itoa(period), func(t *testing.T) {
			res, _ := testClose.MovingAverage(fn, period)
			ma := fn(period)
			cmp := testClose.Map(ma.Update, false).Slice(-res.Len(), 0)
			if !cmp.Equal(res) {
				t.Log(res)
				t.Log(cmp)
				t.Fatal()
			}
			compare(t, res, "result = talib.%s(testClose, %d)", name, period)
		})
	}
}

func benchMA(b *testing.B, name string, step int, fn MovingAverageFunc) {
	b.Helper()
	b.RunParallel(func(pb *testing.PB) {
		var sink interface{}
		for pb.Next() {
			sink, _ = testClose.MovingAverage(fn, step)
		}
		_ = sink
	})
}

func TestBlah(t *testing.T) {
	N := 9
	res, _, _, _ := testClose.MACD(12, 26, N)
	compare(t, res.Slice(-5, 0), "result, _, _ = talib.%s(testClose, %v)", "MACD", "12, 26, 9")
}
