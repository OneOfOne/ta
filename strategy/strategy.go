package strategy

import (
	"log"

	"go.oneofone.dev/ta"
	"go.oneofone.dev/ta/decimal"
)

type Decimal = decimal.Decimal

type Engine interface {
	Start(onBuy, onSell func() (shares int, pricePerShare Decimal))
	Stop() (shares int, pricePershare, availableBalance Decimal)
}

type Tick struct {
	Price  Decimal
	Volume int
}

type Strategy interface {
	Update(*Tick) (buy, sell bool)
}

type Tx struct {
	initial   Decimal
	Value     Decimal
	LastPrice Decimal
	Bought    int
	Sold      int
	Shorted   int
	Held      int
}

func (t *Tx) Total() Decimal {
	return t.Value + (Decimal(t.Held) * t.LastPrice)
}

// PL - Profit / Loss
func (t *Tx) PL() Decimal {
	return t.Total() - t.initial
}

// PLPerc - Profit/Loss percent
func (t *Tx) PLPerc() Decimal {
	return ((t.PL() / t.Total()) * 100).Floor(100)
}

func ApplySlice(acc Account, str Strategy, symbol string, data *ta.TA) *Tx {
	inp := make(chan *Tick, 1)
	go func() {
		for i := 0; i < data.Len(); i++ {
			inp <- &Tick{Price: data.Get(i)}
		}
		close(inp)
	}()
	var last *Tx
	for t := range Apply(acc, str, symbol, inp) {
		last = &t
	}
	return last
}

func Apply(acc Account, str Strategy, symbol string, src <-chan *Tick) <-chan Tx {
	ch := make(chan Tx, len(src))
	go func() {
		defer close(ch)
		initial, _, _ := acc.Balance()
		tx := Tx{
			initial: initial,
			Held:    acc.Shares(symbol),
		}
		for t := range src {
			shouldBuy, shouldSell := str.Update(t)
			if tx.LastPrice == 0 {
				tx.Value = tx.initial + (Decimal(tx.Held) * t.Price)
			}
			tx.LastPrice = t.Price
			if shouldBuy && shouldSell {
				log.Printf("[strategy] %T.Update() returned both buy and sell", str)
				shouldBuy = false
			}

			if shouldBuy {
				shares, pricePerShare := acc.Buy(symbol, t.Price)
				if shares == 0 {
					continue
				}
				tx.Bought += shares
				tx.Held += shares
				tx.Value -= Decimal(shares) * pricePerShare
				select {
				case ch <- tx:
				default:
				}
			}

			if shouldSell {
				shares, pricePerShare := acc.Sell(symbol, t.Price)
				if shares == 0 {
					continue
				}
				tx.Sold += shares
				tx.Held -= shares
				tx.Value += Decimal(shares) * pricePerShare
				select {
				case ch <- tx:
				default:
				}
			}
		}

		select {
		case ch <- tx:
		default:
		}
	}()
	return ch
}
