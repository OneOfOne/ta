package ta

import (
	"testing"

	"go.oneofone.dev/ta/decimal"
)

func TestBlackScholes(t *testing.T) {
	// this is pretty much translated from https://github.com/MattL922/black-scholes/blob/master/test/black-scholes.js
	t.Run("t>0, v>0", func(t *testing.T) {
		if BlackScholes(30, 34, .25, .2, .08, true).NotEqual(0.23834902311961947) {
			t.Fatal("should return a call price of 0.23834902311961947")
		}

		if BlackScholes(30, 34, .25, .2, .08, false).NotEqual(3.5651039155492974) {
			t.Fatal("should return a put price of 3.5651039155492974")
		}
	})
	t.Run("t>0, v=0, out-of-the-money", func(t *testing.T) {
		if BlackScholes(30, 34, .25, 0, .08, true).NotEqual(0) {
			t.Fatal("should return a call price of 0")
		}

		if BlackScholes(35, 34, .25, 0, .08, false).NotEqual(0) {
			t.Fatal("should return a put price of 0")
		}
	})
	t.Run("t=0, v>0, out-of-the-money", func(t *testing.T) {
		if BlackScholes(30, 34, 0, 0.1, .08, true).NotEqual(0) {
			t.Fatal("should return a call price of 0")
		}

		if BlackScholes(35, 34, 0, 0.1, .08, false).NotEqual(0) {
			t.Fatal("should return a put price of 0")
		}
	})
	t.Run("t=0, v=0, out-of-the-money", func(t *testing.T) {
		if BlackScholes(30, 34, 0, 0, .08, true).NotEqual(0) {
			t.Fatal("should return a call price of 0")
		}

		if BlackScholes(35, 34, 0, 0, .08, false).NotEqual(0) {
			t.Fatal("should return a put price of 0")
		}
	})
	// It may seem odd that the call is worth significantly more than the put when
	// they are both $2 in the money.  This is because the call theoretically has
	// unlimited profit potential.  The put can only make money until the underlying
	// goes to zero.  Therefore the call has more value.
	t.Run("t>0,v=0, in-the-money", func(t *testing.T) {
		if BlackScholes(36, 34, .25, 0, .08, true).NotEqual(2.673245107570324) {
			t.Fatal("should return a call price of 2.673245107570324")
		}

		if BlackScholes(32, 34, .25, 0, .08, false).NotEqual(1.3267548924296761) {
			t.Fatal("should return a put price of 1.3267548924296761")
		}
	})
	t.Run("t=0,v>0, in-the-money", func(t *testing.T) {
		if BlackScholes(36, 34, 0, 0.1, .08, true).NotEqual(2) {
			t.Fatal("should return a call price of 2")
		}

		if BlackScholes(32, 34, 0, 0.1, .08, false).NotEqual(2) {
			t.Fatal("should return a put price of 2")
		}
	})
	t.Run("t=0, v=0, in-the-money", func(t *testing.T) {
		if BlackScholes(36, 34, 0, 0, .08, true).NotEqual(2) {
			t.Fatal("should return a call price of 2")
		}

		if BlackScholes(32, 34, 0, 0, .08, false).NotEqual(2) {
			t.Fatal("should return a put price of 2")
		}
	})
	t.Run("Standard Normal Cumulative Distribution Function", func(t *testing.T) {
		if CDF(0).NotEqual(.5) {
			t.Fatal("should return 0.5")
		}
		if CDF(decimal.Inf).NotEqual(1) {
			t.Fatal("should return 1")
		}
		if CDF(decimal.NegInf).NotEqual(0) {
			t.Fatal("should return 0")
		}
		if (CDF(1) - CDF(-1)).NotEqual(0.6826894921370861) {
			t.Fatal("should return 1 standard deviation")
		}
		if (CDF(2) - CDF(-2)).NotEqual(0.9544997361036414) {
			t.Fatal("should return 2 standard deviations")
		}
		if (CDF(3) - CDF(-3)).NotEqual(0.99730020393674) {
			t.Fatal("should return 3 standard deviations")
		}
	})
	t.Run("Ω", func(t *testing.T) {
		if Omega(30, 34, .25, .2, .08).NotEqual(-1.00163142954006) {
			t.Fatal("should return -1.00163142954006")
		}
	})
}

func TestIV(t *testing.T) {
	iv := ImpliedVolatility(2, 101, 100, 0.1, .0015, true)
	if iv.NotEqual(0.11406250000000001) {
		t.Fatal("call iv should be ~.11")
	}
	iv = ImpliedVolatility(2, 101, 100, 0.1, .0015, false)
	if iv.NotEqual(0.1953125) {
		t.Fatal("put iv should be ~.19")
	}
}
