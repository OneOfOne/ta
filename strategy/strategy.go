package strategy

import (
	"log"

	"go.oneofone.dev/ta/csvticks"
	"go.oneofone.dev/ta/decimal"
)

type Decimal = decimal.Decimal

type Engine interface {
	Start(onBuy, onSell func() (shares int, pricePerShare Decimal))
	Stop() (shares int, pricePershare, availableBalance Decimal)
}

type Candle struct {
	Open   Decimal
	High   Decimal
	Low    Decimal
	Close  Decimal
	Volume int
}

type Strategy interface {
	Setup(candles []*Candle)
	Update(*Candle) (buy, sell bool)
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

func ApplySlice(acc Account, str Strategy, symbol string, data csvticks.Ticks) *Tx {
	inp := make(chan *Candle, 1)
	go func() {
		for _, t := range data {
			inp <- &Candle{Close: t.Close, Volume: int(t.Volume), Open: t.Open, High: t.High, Low: t.Low}
		}
		close(inp)
	}()
	var last *Tx
	for t := range Apply(acc, str, symbol, inp) {
		last = &t
	}
	return last
}

func Apply(acc Account, str Strategy, symbol string, src <-chan *Candle) <-chan Tx {
	ch := make(chan Tx, len(src))
	go func() {
		defer close(ch)
		initial, _, _ := acc.Balance()
		tx := Tx{
			initial: initial,
			Held:    acc.Shares(symbol),
		}
		for c := range src {
			shouldBuy, shouldSell := str.Update(c)
			if tx.LastPrice == 0 {
				tx.Value = tx.initial + (Decimal(tx.Held) * c.Close)
			}
			tx.LastPrice = c.Close
			if shouldBuy && shouldSell {
				log.Printf("[strategy] %v.Update() returned both buy and sell", str)
				if tx.Held > 0 {
					shouldBuy = false
				} else {
					shouldSell = false
				}

			}

			if shouldBuy {
				shares, pricePerShare := acc.Buy(symbol, c.Close)
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
				shares, pricePerShare := acc.Sell(symbol, c.Close)
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
