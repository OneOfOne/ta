package strategy

import (
	"go.oneofone.dev/ta"
	"go.oneofone.dev/ta/decimal"
)

func RSI(period, oversold, overbought int) Strategy {
	return &rsi{
		rsi:        ta.RSI(period),
		oversold:   Decimal(oversold),
		overbought: Decimal(overbought),
		idx:        period,
	}
}

type rsi struct {
	rsi        ta.Study
	oversold   Decimal
	overbought Decimal
	last       Decimal
	idx        int
	dir        int8
}

func (r *rsi) Update(t *Tick) (buy, sell bool) {
	v := t.Price
	v = r.rsi.Update(v)
	switch {
	case r.idx > 0:
		r.idx--
	case decimal.Crosover(v, r.last, r.overbought):
		r.dir = 1
	case decimal.Crossunder(v, r.last, r.oversold):
		r.dir = -1
	default:
		r.dir = 0
	}
	r.last = v

	return r.dir > 0, r.dir < 0
}

func MACD(fastPeriod, slowPeriod, signalPeriod int) Strategy {
	return &macd{
		macd: ta.MACD(fastPeriod, slowPeriod, signalPeriod),
		idx:  (fastPeriod + slowPeriod + signalPeriod) / 3,
	}
}

func MACDWithResistance(resistance, fastPeriod, slowPeriod, signalPeriod int) Strategy {
	if resistance < 0 {
		resistance = 0
	}
	return &macd{
		macd: ta.MACD(fastPeriod, slowPeriod, signalPeriod),
		idx:  (fastPeriod + slowPeriod + signalPeriod) / 3,
		res:  resistance,
	}
}

func MACDExt(resistance, fastPeriod, slowPeriod, signalPeriod int, fn ta.MovingAverageFunc) Strategy {
	if resistance < 0 {
		resistance = 0
	}
	return &macd{
		macd: ta.MACDExt(fastPeriod, slowPeriod, signalPeriod, fn),
		idx:  (fastPeriod + slowPeriod + signalPeriod) / 3,
		res:  resistance,
	}
}

func MACDMulti(resistance int, fast, slow, signal ta.MovingAverage) Strategy {
	if resistance < 0 {
		resistance = 0
	}
	return &macd{
		macd: ta.MACDMulti(fast, slow, signal),
		idx:  fast.Len() + slow.Len() + signal.Len()/3,
		res:  resistance,
	}
}

type macd struct {
	macd ta.Study
	last Decimal
	res  int
	idx  int
	dir  int
}

func (r *macd) Update(t *Tick) (buy, sell bool) {
	v := r.macd.Update(t.Price)

	switch {
	case r.idx > 0:
		r.idx--
	case decimal.Crossunder(v, r.last, 0):
		r.dir = -1
	case decimal.Crosover(v, r.last, 0):
		r.dir = 1
	case v > r.last && r.dir > 0:
		r.dir++
	case v < r.last && r.dir < 0:
		r.dir--
	default:
		r.dir = 0
	}

	r.last = v
	return r.dir > r.res, r.dir < -r.res
}
