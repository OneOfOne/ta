package strategy

import (
	"context"
	"math"
	"strconv"

	"go.oneofone.dev/ta"
	"go.oneofone.dev/ta/decimal"
)

type Decimal = decimal.Decimal

type Strategy interface {
	Update(v Decimal)
	ShouldBuy() bool
	ShouldSell() bool
}

type Result struct {
	Symbol       string
	StartBalance Decimal
	Balance      Decimal

	Orders []*Order

	Bought int
	Sold   int

	LastPrice Decimal
}

func (r *Result) Total() Decimal {
	return r.Balance + r.SharesValue()
}

func (r *Result) NumShares() (n int) {
	for _, o := range r.Orders {
		n += o.Count
	}
	return
}

func (r *Result) SharesValue() (n Decimal) {
	for _, o := range r.Orders {
		n += (o.Price * Decimal(o.Count))
	}
	return
}

// PL - Profit / Loss
func (r *Result) PL() Decimal { return r.Total() - r.StartBalance }

// PLPerc - Profit/Loss percent
func (r *Result) PLPerc() Decimal {
	return ((r.PL() / r.Total()) * 100).Floor(100)
}

type (
	Order struct {
		ID     string
		Symbol string
		Price  Decimal
		Count  int
	}

	Options struct {
		Balance         Decimal
		Orders          []*Order
		MaxSharesToHold int

		AllowShort bool

		Buy  func(r *Result) (orders []*Order)
		Sell func(r *Result, o Order) (pricePerShare Decimal)
	}
)

func ApplyLive(ctx context.Context, str Strategy, symbol string, input <-chan Decimal, opts Options) <-chan *Result {
	res := make(chan *Result, 1)
	if opts.Balance < 1 {
		panic("balance < 1")
	}

	if opts.MaxSharesToHold == 0 {
		opts.MaxSharesToHold = int(math.MaxInt64)
	}

	done := ctx.Done()
	go func() {
		r := &Result{
			Symbol:       symbol,
			StartBalance: opts.Balance,
			Balance:      opts.Balance,
			Orders:       append([]*Order(nil), opts.Orders...),
		}

		defer func() {
			res <- r
			close(res)
		}()

	L:
		select {
		case v, ok := <-input:
			if !ok {
				return
			}
			r.LastPrice = v
			str.Update(v)

			shouldBuy := str.ShouldBuy() && r.Balance > v && r.NumShares() < opts.MaxSharesToHold
			shouldSell := str.ShouldSell() && (len(r.Orders) > 0 || opts.AllowShort)

			if shouldBuy && shouldSell {
				shouldBuy = false
			}

			switch {
			case shouldBuy:
				bought := opts.Buy(r)
				if len(bought) == 0 {
					goto L
				}
				for _, o := range bought {
					r.Bought += o.Count
					r.Balance -= (o.Price * Decimal(o.Count))
				}
				r.Orders = append(r.Orders, bought...)

			case shouldSell:
				out := r.Orders[:0]
				for _, o := range r.Orders {
					pricePerShare := opts.Sell(r, *o)
					if pricePerShare == 0 {
						out = append(out, o)
						continue
					}
					r.Sold += o.Count
					r.Balance += (pricePerShare * Decimal(o.Count))
				}
				r.Orders = out
			default:

			}
		case <-done:
			return
		}
		goto L
	}()
	return res
}

func Apply(str Strategy, symbol string, data *ta.TA, startBalance float64, maxShares int) *Result {
	in := make(chan Decimal, 10)
	go func() {
		for i := 0; i < data.Len(); i++ {
			in <- data.Get(i)
		}
		close(in)
	}()

	id := 0

	return <-ApplyLive(context.Background(), str, symbol, in, Options{
		Balance:         Decimal(startBalance),
		MaxSharesToHold: maxShares,

		Buy: func(r *Result) (out []*Order) {
			id++
			numShares := ta.MinInt(int(r.Balance/r.LastPrice), maxShares-r.NumShares())
			if numShares == 0 {
				return
			}
			return []*Order{
				{ID: strconv.Itoa(id), Count: numShares, Price: r.LastPrice},
			}
		},
		Sell: func(r *Result, o Order) (pricePerShare Decimal) {
			return r.LastPrice
		},
	})
}

func Merge(matchAll bool, strats ...Strategy) Strategy {
	if len(strats) < 2 {
		return strats[0]
	}
	return &merge{strats: strats, all: matchAll}
}

type merge struct {
	strats []Strategy
	all    bool
}

func (m *merge) Update(v Decimal) {
	for _, s := range m.strats {
		s.Update(v)
	}
}

func (m *merge) ShouldBuy() bool {
	for _, s := range m.strats {
		if s.ShouldBuy() {
			if !m.all {
				return true
			}
		} else if m.all {
			return false
		}
	}
	return true
}

func (m *merge) ShouldSell() bool {
	for _, s := range m.strats {
		if s.ShouldSell() {
			if !m.all {
				return true
			}
		} else if m.all {
			return false
		}
	}
	return true
}

func Mixed(buyStrat, sellStrat Strategy) Strategy {
	return &mixed{buyStrat, sellStrat}
}

type mixed [2]Strategy

func (m *mixed) Update(v Decimal) {
	m[0].Update(v)
	m[1].Update(v)
}

func (m *mixed) ShouldBuy() bool {
	return m[0].ShouldBuy()
}

func (m *mixed) ShouldSell() bool {
	return m[1].ShouldSell()
}
