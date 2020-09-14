package strategy

import (
	"sync"

	"go.oneofone.dev/ta/decimal"
)

func NewAccount(id string, opts AccountOptions) Account {
	a := account{
		id:     id,
		opts:   opts,
		bp:     opts.BuyingPower,
		shares: make(map[string]shares, len(opts.Shares)),
	}

	for sym, cnt := range opts.Shares {
		a.shares[sym] = shares{count: cnt}
	}

	return &a
}

type Account interface {
	ID() string
	Balance() (buyingPower, onHold, sharesValue Decimal)
	Shares(symbol string) int
	Buy(symbol string, price Decimal) (shares int, pricePerShare Decimal)
	Sell(symbol string, price Decimal) (shares int, pricePerShare Decimal)
}

type shares struct {
	count        int
	lastPrice    Decimal
	lastBuyPrice Decimal
}

type (
	TxFunc = func(symbol string, maxShares int, price Decimal) (shares int, pricePerShare Decimal)

	AccountOptions struct {
		BuyingPower        Decimal
		MaxSharesPerSymbol int

		Shares    map[string]int
		CanShort  bool
		ReuseCash bool // implies day trading with margin

		BuyFunc  TxFunc
		SellFunc TxFunc
	}
)

type account struct {
	id     string
	mux    sync.RWMutex
	opts   AccountOptions
	bp     Decimal
	onHold Decimal
	shares map[string]shares
}

func (a *account) ID() string { return a.id }
func (a *account) Balance() (buyingPower, onHold, sharesValue Decimal) {
	a.mux.RLock()
	defer a.mux.RUnlock()
	for _, sh := range a.shares {
		sharesValue += sh.lastPrice * Decimal(sh.count)
	}
	return a.bp, a.onHold, sharesValue
}

func (a *account) Shares(sym string) int {
	a.mux.RLock()
	defer a.mux.RUnlock()
	return a.shares[sym].count
}

func (a *account) Buy(symbol string, price Decimal) (shares int, pricePerShare Decimal) {
	a.mux.Lock()
	defer a.mux.Unlock()

	if price > a.bp {
		return
	}

	sh, ok := a.shares[symbol]
	if ok {
		sh.lastPrice = price
		a.shares[symbol] = sh
	}

	if sh.count >= a.opts.MaxSharesPerSymbol {
		return
	}

	max := a.opts.MaxSharesPerSymbol - sh.count
	for price.Muli(max) > a.bp {
		max--
	}

	if max == 0 {
		return
	}

	if fn := a.opts.BuyFunc; fn != nil {
		shares, pricePerShare = fn(symbol, max, price)
	} else {
		shares, pricePerShare = a.buyAuto(symbol, max, price)
	}

	if shares == 0 {
		return
	}

	a.bp -= (Decimal(shares) * pricePerShare)

	// if a.opts.ReuseCash {

	// } else if sh.count < 0 {
	// 	nsh := shares + sh.count
	// 	if nsh < 1 {
	// 		a.onHold -= (Decimal(shares) * pricePerShare)
	// 	} else {
	// 		a.onHold -= (Decimal(-sh.count) * pricePerShare)
	// 		a.bp -= (Decimal(nsh) * pricePerShare)
	// 	}
	// }

	sh.count += shares
	sh.lastBuyPrice = pricePerShare
	a.shares[symbol] = sh
	return
}

func (a *account) Sell(symbol string, price Decimal) (shares int, pricePerShare Decimal) {
	a.mux.Lock()
	defer a.mux.Unlock()

	sh := a.shares[symbol]
	if sh.count == 0 && !a.opts.CanShort {
		return
	}

	if a.opts.CanShort && -sh.count >= a.opts.MaxSharesPerSymbol {
		return
	}

	max := decimal.MinInt(sh.count, a.opts.MaxSharesPerSymbol)

	if max == 0 {
		return
	}

	if fn := a.opts.BuyFunc; fn != nil {
		shares, pricePerShare = fn(symbol, max, price)
	} else {
		shares, pricePerShare = a.sellAuto(symbol, max, price)
	}

	if shares == 0 {
		return
	}

	if a.opts.ReuseCash {
		a.bp += (Decimal(shares) * pricePerShare)
	} else {
		a.onHold += (Decimal(shares) * pricePerShare)
	}
	sh.count -= shares
	sh.lastBuyPrice = pricePerShare
	a.shares[symbol] = sh

	return
}

func (a *account) buyAuto(symbol string, max int, price Decimal) (shares int, pricePerShare Decimal) {
	shares = max
	pricePerShare = price
	return
}

func (a *account) sellAuto(symbol string, max int, price Decimal) (shares int, pricePerShare Decimal) {
	shares = max
	pricePerShare = price
	return
}
