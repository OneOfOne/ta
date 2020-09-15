package ta

import (
	"sync"
	"testing"
)

func wrapMA(ma MovingAverageFunc) func(p int) Study {
	return func(p int) Study {
		return ma(p)
	}
}
func TestSMA(t *testing.T)  { testMA(t, "SMA", SMA, -1) }
func TestEMA(t *testing.T)  { testMA(t, "EMA", EMA, -1) }
func TestWMA(t *testing.T)  { testMA(t, "WMA", WMA, -1) }
func TestDEMA(t *testing.T) { testMA(t, "DEMA", DEMA, 36) }
func TestTEMA(t *testing.T) { testMA(t, "TEMA", TEMA, 32) }

func TestRSI(t *testing.T) { testStudy(t, "RSI", RSI, -1) }
func TestVar(t *testing.T) {
	vari := func(period int) Study {
		return Variance(period)
	}
	testStudy(t, "VAR", vari, -1)
}
func TestStdDev(t *testing.T) { testStudy(t, "STDDEV", StdDev, -1) }
func TestMin(t *testing.T)    { testStudy(t, "MIN", Min, -1) }
func TestMax(t *testing.T)    { testStudy(t, "MAX", Max, -1) }

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

func TestBBands(t *testing.T) {
	t.Parallel()
	tests := &[...]*[2]Decimal{
		{1.0, 1.0},
		{1.0, 2.0},
		{2.0, 1.0},
		{2.0, 2.0},
		{3.0, 4.0},
	}

	var wg sync.WaitGroup
	for _, ts := range tests {
		for _, p := range &[...]int{2, 5, 10, 20} {
			testBBands(t, &wg, p, ts[0], ts[1], SMA, "SMA")
			testBBands(t, &wg, p, ts[0], ts[1], EMA, "EMA")
			testBBands(t, &wg, p, ts[0], ts[1], WMA, "WMA")
			testBBands(t, &wg, p, ts[0], ts[1], DEMA, "DEMA")
			testBBands(t, &wg, p, ts[0], ts[1], TEMA, "TEMA")
		}
	}
	wg.Wait()
}

func TestVWAP(t *testing.T) {
	data := [][2]Decimal{
		{2.5, 268},
		{7.5, 269},
	}

	vwap := VWAP(2)

	var last Decimal
	for _, d := range data {
		last = vwap.Update(d[0], d[1])
	}

	if last != 268.75 {
		t.Fatalf("expected 268.75, got %v", last)
	}

	bvwap, _ := VWAP(2).ToMulti()

	vol := NewSize(2, true).Append(2.5, 7.5)
	prices := NewSize(2, true).Append(268, 269)
	vw := ApplyMultiVarStudy(bvwap, vol, prices)
	t.Log(vw)
	if last = vw[0].Last(); last != 268.75 {
		t.Fatalf("expected 268.75, got %v", last)
	}

	bb := ApplyMultiVarStudy(BBands(10), testClose)
	t.Log(len(bb))
	t.Log(bb[0])
	t.Log(bb[1])
	t.Log(bb[2])
}
