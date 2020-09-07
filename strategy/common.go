package strategy

import "go.oneofone.dev/ta"

func RSI(period, oversold, overbought int) Strategy {
	return &rsi{
		rsi:        ta.RSI(period),
		oversold:   ta.Decimal(oversold),
		overbought: ta.Decimal(overbought),
		idx:        period,
	}
}

type rsi struct {
	rsi        ta.Study
	oversold   ta.Decimal
	overbought ta.Decimal
	last       ta.Decimal
	idx        int
	dir        int8
}

func (r *rsi) Update(v ta.Decimal) {
	v = r.rsi.Update(v)
	switch {
	case r.idx > -1:
		r.idx--
	case v >= r.overbought && r.last < r.overbought:
		r.dir = -1
	case v >= r.oversold && r.last < r.oversold:
		r.dir = 1
	default:
		r.dir = 0
	}
	r.last = v
}

func (r *rsi) ShouldBuy() bool {
	return r.dir > 0
}

func (r *rsi) ShouldSell() bool {
	return r.dir < 0
}

func MACD(fastPeriod, slowPeriod, signalPeriod int) Strategy {
	return &macd{
		macd: ta.MACD(fastPeriod, slowPeriod, signalPeriod),
		idx:  (fastPeriod + slowPeriod + signalPeriod) / 3,
	}
}

func MACDExt(fast, slow, signal ta.Study) Strategy {
	return &macd{
		macd: ta.MACDExt(fast, slow, signal),
		idx:  fast.Len() + slow.Len() + signal.Len()/3,
	}
}

type macd struct {
	macd ta.MACDStudy
	last ta.Decimal
	idx  int
	dir  int8
}

func (r *macd) Update(v ta.Decimal) {
	_, _, v = r.macd.Update(v)

	switch {
	case r.idx > -1:
		r.idx--
	case v < 0 && r.last > 0:
		r.dir = -1
	case v > 0 && r.last < 0:
		r.dir = 1
	default:
		r.dir = 0
	}
	r.last = v
}

func (r *macd) ShouldBuy() bool {
	return r.dir > 0
}

func (r *macd) ShouldSell() bool {
	return r.dir < 0
}
