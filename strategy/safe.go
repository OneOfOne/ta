package strategy

import (
	"go.oneofone.dev/ta/decimal"
)

// Trailing with Strat sets a trailing win/loss stop point based on the strategy
func Trailing(str Strategy, maxLoss, maxWin Decimal) Strategy {
	return &trailing{str: str}
}

type order struct {
	cost Decimal
	high Decimal
	low  Decimal
}

type trailing struct {
	str    Strategy
	orders []Decimal
	last   Decimal
	idx    int
	dir    int8
}

func (r *trailing) checkSell(v Decimal) int {
	for _, o := range r.orders {
		_ = o
	}
	return 0
}

func (r *trailing) Update(v Decimal) {
	r.str.Update(v)
	switch {
	case r.idx > 0:
		r.idx--
	case decimal.Crosover(v, r.last, 0):
		r.dir = 1
	case decimal.Crossunder(v, r.last, 0):
		r.dir = -1
	default:
		r.dir = 0
	}
	r.last = v
}

func (r *trailing) ShouldBuy() bool {
	return r.dir > 0
}

func (r *trailing) ShouldSell() bool {
	return r.dir < 0
}
