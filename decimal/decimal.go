package decimal

import (
	"math"
	"math/big"
	"strconv"

	"gonum.org/v1/gonum/floats/scalar"
)

const Epsilon = 1e-7

var MarshalAsString = true

func Zero() Decimal {
	return 0
}

func One() Decimal {
	return 1
}

func FromInt(v int64) Decimal {
	return Decimal(v)
}

func FromUint(v uint64) Decimal {
	return Decimal(v)
}

func FromFloat(v float64) Decimal {
	return Decimal(v)
}

func FromString(v string) Decimal {
	d, _ := ParseString(v)
	return d
}

func ParseString(v string) (Decimal, error) {
	f, err := strconv.ParseFloat(v, 64)
	return Decimal(f), err
}

func FromBigFloat(v *big.Float) Decimal {
	f, _ := v.Float64()
	return Decimal(f)
}

type Decimal float64

func (d Decimal) Add(x Decimal) Decimal {
	return d + x
}

func (d Decimal) Addf(x float64) Decimal {
	return d.Add(Decimal(x))
}

func (d Decimal) Sub(x Decimal) Decimal {
	return d - x
}

func (d Decimal) Subf(x float64) Decimal {
	return d.Sub(Decimal(x))
}

func (d Decimal) Mulf(x float64) Decimal {
	return d.Mul(Decimal(x))
}

func (d Decimal) Mul(x Decimal) Decimal {
	return d * x
}

func (d Decimal) Divf(x float64) Decimal {
	return d.Div(Decimal(x))
}

func (d Decimal) Div(x Decimal) Decimal {
	return d / x
}

func (d Decimal) Abs() Decimal {
	v := math.Abs(float64(d))
	return Decimal(v)
}

func (d Decimal) Neg() Decimal {
	return -d
}

func (d Decimal) IsNaN() bool {
	return math.IsNaN(d.Float())
}

func (d Decimal) IsInf(sign int) bool {
	return math.IsInf(d.Float(), sign)
}

func (d Decimal) Pow2() Decimal {
	return d.Pow(2)
}

func (d Decimal) Sqrt() Decimal {
	v := math.Sqrt(float64(d))
	return Decimal(v)
}

// Atan uses shopspring/decimal
func (d Decimal) Atan() Decimal {
	v := math.Atan(float64(d))
	return Decimal(v)
}

// Pow uses shopspring/decimal if n < 10, otherwise a simple loop
func (d Decimal) Pow(n int) Decimal {
	v := math.Pow(float64(d), float64(n))
	return Decimal(v)
}

func (d Decimal) Cmp(x Decimal) int {
	if EqualApprox(float64(d), float64(x), Epsilon) {
		return 0
	}
	if d < x {
		return -1
	}
	return 1
}

func (d Decimal) Cmpf(x float64) int {
	return d.Cmp(Decimal(x))
}

func (d Decimal) LessThanOrEqual(x Decimal) bool {
	return d.Cmp(x) <= 0
}

func (d Decimal) LessThanOrEqualf(x float64) bool {
	return d.Cmpf(x) <= 0
}

func (d Decimal) GreaterThanOrEqual(x Decimal) bool {
	return d.Cmp(x) >= 0
}

func (d Decimal) GreaterThanOrEqualf(x float64) bool {
	return d.Cmpf(x) >= 0
}

func (d Decimal) LessThan(x Decimal) bool {
	return d.Cmp(x) < 0
}

func (d Decimal) LessThanf(x float64) bool {
	return d.Cmpf(x) < 0
}

func (d Decimal) GreaterThan(x Decimal) bool {
	return d.Cmp(x) > 0
}

func (d Decimal) GreaterThanf(x float64) bool {
	return d.Cmpf(x) > 0
}

func (d Decimal) Equal(x Decimal) bool {
	return d.Cmp(x) == 0
}

func (d Decimal) Equalf(x float64) bool {
	return d.Cmpf(x) == 0
}

func (d Decimal) NotEqual(x Decimal) bool {
	return !d.Equal(x)
}

func (d Decimal) NotEqualf(x float64) bool {
	return !d.Equalf(x)
}

func (d Decimal) Big() *big.Float {
	return big.NewFloat(float64(d))
}

func (d Decimal) Float() float64 {
	return float64(d)
}

func (d Decimal) MarshalJSON() ([]byte, error) {
	if MarshalAsString {
		return []byte(`"` + d.String() + `"`), nil
	}
	return []byte(d.String()), nil
}

func (d Decimal) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Decimal) UnmarshalJSON(p []byte) error {
	return d.UnmarshalText(p)
}

func (d *Decimal) UnmarshalText(p []byte) (err error) {
	if len(p) == 0 {
		return nil
	}
	if len(p) > 2 && p[0] == '"' && p[len(p)-1] == '"' {
		p = p[1 : len(p)-1]
	}

	*d, err = ParseString(string(p))
	return
}

func (d Decimal) String() string {
	return d.Text('g', 20)
}

func (d Decimal) Text(fmt byte, prec int) string {
	return strconv.FormatFloat(d.Float(), fmt, prec, 64)
}

func EqualApprox(a, b, epsilon float64) bool {
	return scalar.EqualWithinAbsOrRel(a, b, epsilon, epsilon)
}
