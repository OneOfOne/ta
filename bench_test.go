package ta

import "testing"

func BenchmarkSMA(b *testing.B)  { benchMA(b, "SMA", 10, SMA) }
func BenchmarkEMA(b *testing.B)  { benchMA(b, "EMA", 10, EMA) }
func BenchmarkWMA(b *testing.B)  { benchMA(b, "WMA", 10, WMA) }
func BenchmarkDEMA(b *testing.B) { benchMA(b, "DEMA", 10, DEMA) }
func BenchmarkTEMA(b *testing.B) { benchMA(b, "TEMA", 10, TEMA) }

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
