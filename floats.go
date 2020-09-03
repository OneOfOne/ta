package ta

import "math"

func FromFloats(vs []float64, transformers ...func(float64) float64) TA {
	ds := make(TA, 0, len(vs))
	for _, v := range vs {
		for _, fn := range transformers {
			v = fn(v)
		}
		ds = append(ds, F(v))
	}
	return ds
}

// FloatsSMA - Simple Moving Average
func FloatsSMA(data []float64, period int) []float64 {
	out, _ := FromFloats(data).SMA(period)
	return out.Floats()
}

// FloatsEMA - Exponential Moving Average
func FloatsEMA(data []float64, period int) []float64 {
	out, _ := FromFloats(data).EMA(period)
	return out.Floats()
}

// FloatsWMA - Weighted Moving Average
func FloatsWMA(data []float64, period int) []float64 {
	out, _ := FromFloats(data).WMA(period)
	return out.Floats()
}

// FloatsDEMA - Double Exponential Moving Average
func FloatsDEMA(data []float64, period int) []float64 {
	out, _ := FromFloats(data).DEMA(period)
	return out.Floats()
}

// FloatsTEMA - Triple Exponential Moving Average
func FloatsTEMA(data []float64, period int) []float64 {
	out, _ := FromFloats(data).TEMA(period)
	return out.Floats()
}

// FloatsTMAma - Triangular Moving Average
// func FloatsTriMA(data []float64, period int) []float64 {
// 	return FromFloats(data).Trima(period).Floats()
// }

// // Kama - Kaufman Adaptive Moving Average
// func Kama(data []float64, period int) []float64 {
// 	return FromFloats(data).Kama(period).Floats()
// }

// // Mama - MESA Adaptive Moving Average (lookback=32)
// func Mama(data []float64, fastLimit, slowLimit float64) ([]float64, []float64) {
// 	m, f := FromFloats(data).Mama(32, fastLimit, slowLimit)
// 	return m.Floats(), f.Floats()
// }

// // T3 - Triple Exponential Moving Average (T3) (lookback=6*period)
// func T3(data []float64, period int, vfactor float64) []float64 {
// 	return FromFloats(data).T3(period, 6, vfactor).Floats()
// }

/* Math Transform Functions */

// Acos - Vector Trigonometric ACOS
func Acos(data []float64) []float64 {
	return FromFloats(data, math.Acos).Floats()
}

// Asin - Vector Trigonometric ASIN
func Asin(data []float64) []float64 {
	return FromFloats(data, math.Asin).Floats()
}

// Atan - Vector Trigonometric ATAN
func Atan(data []float64) []float64 {
	return FromFloats(data, math.Atan).Floats()
}

// Ceil - Vector CEIL
func Ceil(data []float64) []float64 {
	return FromFloats(data, math.Ceil).Floats()
}

// Cos - Vector Trigonometric COS
func Cos(data []float64) []float64 {
	return FromFloats(data, math.Cos).Floats()
}

// Cosh - Vector Trigonometric COSH
func Cosh(data []float64) []float64 {
	return FromFloats(data, math.Cosh).Floats()
}

// Exp - Vector arithmetic EXP
func Exp(data []float64) []float64 {
	return FromFloats(data, math.Exp).Floats()
}

// Floor - Vector FLOOR
func Floor(data []float64) []float64 {
	return FromFloats(data, math.Floor).Floats()
}

// Ln - Vector natural log LN
func Ln(data []float64) []float64 {
	return FromFloats(data, math.Log).Floats()
}

// Log10 - Vector LOG10
func Log10(data []float64) []float64 {
	return FromFloats(data, math.Log10).Floats()
}

// Sin - Vector Trigonometric SIN
func Sin(data []float64) []float64 {
	return FromFloats(data, math.Sin).Floats()
}

// Sinh - Vector Trigonometric SINH
func Sinh(data []float64) []float64 {
	return FromFloats(data, math.Sinh).Floats()
}

// Sqrt - Vector SQRT
func Sqrt(data []float64) []float64 {
	return FromFloats(data, math.Sqrt).Floats()
}

// Tan - Vector Trigonometric TAN
func Tan(data []float64) []float64 {
	return FromFloats(data, math.Tan).Floats()
}

// Tanh - Vector Trigonometric TANH
func Tanh(data []float64) []float64 {
	return FromFloats(data, math.Tanh).Floats()
}

// Other

// // Var - Variance
// func Var(data []float64, period int) []float64 {
// 	return FromFloats(data).Variance(period).Floats()
// }

// // StdDev - Standard Deviation
// func StdDev(data []float64, period int) []float64 {
// 	return FromFloats(data).StdDev(period).Floats()
// }
