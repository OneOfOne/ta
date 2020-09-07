package strategy

import (
	"context"
	"math"

	"go.oneofone.dev/ta"
)

type Strategy interface {
	Update(v ta.Decimal)
	ShouldBuy() bool
	ShouldSell() bool
}

type Result struct {
	Symbol       string
	StartBalance ta.Decimal
	Balance      ta.Decimal
	SharesValue  ta.Decimal

	Bought int
	Sold   int
	Shares int

	LastPrice ta.Decimal
}

func (r *Result) Total() ta.Decimal    { return r.Balance + r.SharesValue }
func (r *Result) GainLoss() ta.Decimal { return r.Total() - r.StartBalance }
func (r *Result) GainLossPercent() ta.Decimal {
	return ((r.GainLoss() / r.Total()) * 100).Floor(100)
}

type (
	TxFunc func(r *Result) (numShares int, cost ta.Decimal)

	Options struct {
		Balance         ta.Decimal
		Shares          int
		MaxSharesToHold int

		AllowShort bool

		Buy  TxFunc
		Sell TxFunc
	}
)

func ApplyLive(ctx context.Context, str Strategy, symbol string, input <-chan ta.Decimal, opts Options) <-chan *Result {
	res := make(chan *Result, 1)
	if opts.Balance < 1 {
		panic("balance < 1")
	}

	if opts.MaxSharesToHold == 0 {
		opts.MaxSharesToHold = int(math.MaxInt64)
	}

	done := ctx.Done()
	go func() {
		var (
			r = &Result{
				Symbol:       symbol,
				StartBalance: opts.Balance,
				Balance:      opts.Balance,
				Shares:       opts.Shares,
			}
			num          int
			costPerShare ta.Decimal
			tc           ta.Decimal
		)

		defer func() {
			r.SharesValue = ta.Decimal(r.Shares) * r.LastPrice
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

			shouldSell := str.ShouldSell() && (r.Shares > 0 || opts.AllowShort)
			shouldBuy := str.ShouldBuy() && r.Balance > v && r.Shares < opts.MaxSharesToHold

			switch {
			case shouldBuy:
				if r.Balance < v || r.Shares >= opts.MaxSharesToHold {
					break
				}
				if num, costPerShare = opts.Buy(r); num == 0 {
					break
				}
				tc = ta.Decimal(num) * costPerShare
				r.Shares += num
				r.Bought += num
				r.Balance -= tc
				r.SharesValue = ta.Decimal(r.Shares) * v
			case shouldSell:
				if r.Shares == 0 && !opts.AllowShort {
					break
				}
				if num, costPerShare = opts.Sell(r); num == 0 {
					break
				}
				tc = ta.Decimal(num) * costPerShare
				r.Shares -= num
				r.Sold += num
				r.Balance += tc
				r.SharesValue = ta.Decimal(r.Shares) * v
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
	in := make(chan ta.Decimal, 10)
	go func() {
		for i := 0; i < data.Len(); i++ {
			in <- data.At(i)
		}
		close(in)
	}()

	return <-ApplyLive(context.Background(), str, symbol, in, Options{
		Balance:         ta.Decimal(startBalance),
		MaxSharesToHold: maxShares,

		Buy: func(r *Result) (numShares int, costPerShare ta.Decimal) {
			costPerShare = r.LastPrice
			if numShares = ta.MinInt(int(r.Balance/costPerShare), maxShares-r.Shares); numShares == 0 {
				return
			}
			return
		},
		Sell: func(r *Result) (numShares int, costPerShare ta.Decimal) {
			costPerShare = r.LastPrice
			if numShares = r.Shares; numShares < 0 {
				return
			}
			return
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

func (m *merge) Update(v ta.Decimal) {
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

type mixed struct {
	buy, sell Strategy
}

func (m *mixed) Update(v ta.Decimal) {
	m.buy.Update(v)
	m.sell.Update(v)
}

func (m *mixed) ShouldBuy() bool {
	return m.buy.ShouldBuy()
}

func (m *mixed) ShouldSell() bool {
	return m.sell.ShouldSell()
}
