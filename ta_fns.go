package ta

import "math"

// HLC3 - alias for Average(high, low, close)
func HLC3(high, low, close *TA) *TA {
	return Average(high, low, close)
}

// Average - returns the average of the passed in ta's
func Average(tas ...*TA) *TA {
	ln := tas[0].Len()
	for i := 1; i < len(tas); i++ {
		if n := tas[i].Len(); n > ln {
			ln = n
		}
	}

	out := NewSize(ln, true)
	for i := 0; i < ln; i++ {
		var v Decimal
		for _, ta := range tas {
			if i < ta.Len() {
				v = v.Add(ta.Get(i))
			}
		}
		out.Append(v / Decimal(ln))
	}

	return out
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
func (ta *TA) Acos() *TA {
	return ta.Mapf(math.Acos, false)
}

// Asin - Vector Trigonometric ASIN
func (ta *TA) Asin() *TA {
	return ta.Mapf(math.Asin, false)
}

// Atan - Vector Trigonometric ATAN
func (ta *TA) Atan() *TA {
	return ta.Mapf(math.Atan, false)
}

// Ceil - Vector CEIL
func (ta *TA) Ceil() *TA {
	return ta.Mapf(math.Ceil, false)
}

// Cos - Vector Trigonometric COS
func (ta *TA) Cos() *TA {
	return ta.Mapf(math.Cos, false)
}

// Cosh - Vector Trigonometric COSH
func (ta *TA) Cosh() *TA {
	return ta.Mapf(math.Cosh, false)
}

// Exp - Vector arithmetic EXP
func (ta *TA) Exp() *TA {
	return ta.Mapf(math.Exp, false)
}

// Floor - Vector FLOOR
func (ta *TA) Floor() *TA {
	return ta.Mapf(math.Floor, false)
}

// Ln - Vector natural log LN
func (ta *TA) Ln() *TA {
	return ta.Mapf(math.Log, false)
}

// Log10 - Vector LOG10
func (ta *TA) Log10() *TA {
	return ta.Mapf(math.Log10, false)
}

// Sin - Vector Trigonometric SIN
func (ta *TA) Sin() *TA {
	return ta.Mapf(math.Sin, false)
}

// Sinh - Vector Trigonometric SINH
func (ta *TA) Sinh() *TA {
	return ta.Mapf(math.Sinh, false)
}

// Sqrt - Vector SQRT
func (ta *TA) Sqrt() *TA {
	return ta.Mapf(math.Sqrt, false)
}

// Tan - Vector Trigonometric TAN
func (ta *TA) Tan() *TA {
	return ta.Mapf(math.Tan, false)
}

// Tanh - Vector Trigonometric TANH
func (ta *TA) Tanh() *TA {
	return ta.Mapf(math.Tanh, false)
}

// Other

// // Var - Variance
// func (ta *TA) Var(data []float64, period int) *TA {
// 	return FromFloats(data).Variance(period)
// }

// // StdDev - Standard Deviation
// func StdDev(data []float64, period int) []float64 {
// 	return FromFloats(data).StdDev(period).Floats()
// }
