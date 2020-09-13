package ta

import (
	"math"
	"sync"

	"go.oneofone.dev/ta/decimal"
)

const (
	sqrt2Pi Decimal = 2.506628274631000241612355239340104162693023681640625

	mE Decimal = math.E
)

var (
	factCache   *[100]*[2]Decimal
	optionsOnce sync.Once
)

/*
function stdNormCDF(x)
{
  var probability = 0;
  // avoid divergence in the series which happens around +/-8 when summing the
  // first 100 terms
  if(x >= 8)
  {
    probability = 1;
  }
  else if(x <= -8)
  {
    probability = 0;
  }
  else
  {
    for(var i = 0; i < 100; i++)
    {
      probability += (Math.pow(x, 2*i+1)/_doubleFactorial(2*i+1));
    }
    probability *= Math.pow(Math.E, -0.5*Math.pow(x, 2));
    probability /= Math.sqrt(2*Math.PI);
    probability += 0.5;
  }
  return probability;
}
*/

func initOptions() {
	var c [100]*[2]Decimal
	for i := 0; i < 100; i++ {
		v := 2*Decimal(i) + 1
		c[i] = &[2]Decimal{
			v,
			doubleFact(v),
		}
	}
	factCache = &c
}

func doubleFact(v Decimal) Decimal {
	f := Decimal(1)
	for ; v > 1; v -= 2 {
		f *= v
	}
	return f
}

// CDF - Standard normal cumulative distribution function
// The probability is estimated by expanding the CDF into a series using the first 100 terms.
// See https://en.wikipedia.org/wiki/Normal_distribution#Cumulative_distribution_function
//
// x is the upper bound to integrate over. This is P{Z <= x} where Z is a standard normal random variable.
// returns the probability that a standard normal random variable will be less than or equal to x
func CDF(x Decimal) Decimal {
	optionsOnce.Do(initOptions)
	if x >= 8 {
		return 1
	}
	if x <= -8 {
		return 0
	}

	var prob Decimal
	for i := 0; i < 100; i++ {
		c := factCache[i]
		prob += x.Pow(c[0]) / c[1]
	}
	prob *= mE.Pow(-0.5 * x.Pow2())
	prob /= sqrt2Pi
	prob += 0.5
	return prob
}

// /**
//  * Black-Scholes option pricing formula.
//  * See {@link http://en.wikipedia.org/wiki/Black%E2%80%93Scholes_model#Black-Scholes_formula|Wikipedia page}
//  * for pricing puts in addition to calls.
//  *
//  * @param   {Number} s       Current price of the underlying
//  * @param   {Number} k       Strike price
//  * @param   {Number} t       Time to experiation in years
//  * @param   {Number} v       Volatility as a decimal
//  * @param   {Number} r       Anual risk-free interest rate as a decimal
//  * @param   {String} callPut The type of option to be priced - "call" or "put"
//  * @returns {Number}         Price of the option
//  */
//  function blackScholes(s, k, t, v, r, callPut)
//  {
//    var price = null;
//    var w = (r * t + Math.pow(v, 2) * t / 2 - Math.log(k / s)) / (v * Math.sqrt(t));
//    if(callPut === "call")
//    {
// 	 price = s * stdNormCDF(w) - k * Math.pow(Math.E, -1 * r * t) * stdNormCDF(w - v * Math.sqrt(t));
//    }
//    else // put
//    {
// 	 price = k * Math.pow(Math.E, -1 * r * t) * stdNormCDF(v * Math.sqrt(t) - w) - s * stdNormCDF(-w);
//    }
//    return price;
//  }

// BlackScholes option pricing formula for pricing puts and calls
// See https://en.wikipedia.org/wiki/Black%E2%80%93Scholes_model#Black-Scholes_formula
// s current price of the underlying
// k Strike price
// t time to experiation in years (num days / 365)
// v volatility
// r annual risk-free interest rate
// isCall the type of option, true for Call and false for Put
func BlackScholes(s, k, t, v, r Decimal, isCall bool) Decimal {
	Ω := Omega(s, k, t, v, r)
	if isCall {
		return s*CDF(Ω) - k*mE.Pow(-1*r*t)*CDF(Ω-v*t.Sqrt())
	}
	return k*mE.Pow(-1*r*t)*CDF(v*t.Sqrt()-Ω) - s*CDF(-Ω)
}

// Omega - calcuates Ω as defined in the Black-Scholes formula
// s current price of the underlying
// k Strike price
// t time to experiation in years (num days / 365)
// v volatility
// r annual risk-free interest rate
func Omega(s, k, t, v, r Decimal) Decimal {
	return (r*t + v.Pow2()*t/2 - (k / s).Log()) / (v * t.Sqrt())
}

// ImpliedVolatility is an alias for ImpliedVolatilityWithEstimate(expectedCost, s, k, t, r, 0.1, isCall)
func ImpliedVolatility(expectedCost, s, k, t, r Decimal, isCall bool) Decimal {
	return ImpliedVolatilityWithEstimate(expectedCost, s, k, t, r, 0.1, isCall)
}

// ImpliedVolatilityWithEstimate calculates a close estimate of implied volatility given an option price
// A binary search type approach is used to determine the implied volatility
// expectedCost The market price of the option
// s current price of the underlying
// k Strike price
// t time to experiation in years (num days / 365)
// v volatility
// r annual risk-free interest rate
// estimate a initial estimate of implied volatility
// isCall the type of option, true for Call and false for Put
func ImpliedVolatilityWithEstimate(expectedCost, s, k, t, r, estimate Decimal, isCall bool) Decimal {
	low, high := Decimal(0), decimal.Inf
	exp100 := expectedCost * 100
	for i := 0; i < 100; i++ {
		actual := BlackScholes(s, k, t, estimate, r, isCall) * 100
		if exp100 == actual.Floor(1) {
			break
		}
		if actual > exp100 {
			high = estimate
			estimate = (estimate-low)/2 + low
		} else {
			low = estimate
			if estimate = (high-estimate)/2 + estimate; !estimate.IsFinate() {
				estimate = low * 2
			}
		}
	}
	return estimate
}
