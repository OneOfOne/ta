package strategy

import (
	"go.oneofone.dev/ta/decimal"
)

// Trailing with Strat sets a trailing win/loss stop point based on the strategy
func Trailing(str Strategy, maxGain, maxLoss Decimal) Strategy {
	return &trailing{str: str}
}

type trailingOrder struct {
	Order
	max Decimal
	min Decimal
}

type trailing struct {
	dummyStrategy
	str      Strategy
	gainPerc Decimal
	lossPerc Decimal
	orders   []*trailingOrder
	last     Decimal
	idx      int
	dir      int8
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

func (r *trailing) NotifyBuy(o Order) {
	no := &trailingOrder{
		Order: o,
	}
	no.max = o.Price * (1 + r.gainPerc)
	no.min = o.Price * (1 - r.lossPerc)
	r.orders = append(r.orders, no)
}

func (r *trailing) ShouldSell() bool {
	return r.dir < 0
}
