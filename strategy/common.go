package strategy

import (
	"go.oneofone.dev/ta"
	"go.oneofone.dev/ta/decimal"
)

func RSI(period, oversold, overbought int, fn ta.MovingAverageFunc) Strategy {
	r := ta.RSI(period)
	if fn != nil {
		r = ta.RSIExt(fn(period))
	}
	return &rsi{
		rsi:        r,
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

func (s *rsi) Setup(candles []*Candle) {
	for _, c := range candles {
		s.rsi.Update(c.Close)
	}
}

func (s *rsi) Update(c *Candle) (buy, sell bool) {
	v := c.Close
	v = s.rsi.Update(v)
	switch {
	case s.idx > 0:
		s.idx--
	case decimal.Crosover(v, s.last, s.overbought):
		s.dir = 1
	case decimal.Crossunder(v, s.last, s.oversold):
		s.dir = -1
	default:
		s.dir = 0
	}
	s.last = v

	return s.dir > 0, s.dir < 0
}

func MACD(fastPeriod, slowPeriod, signalPeriod int) Strategy {
	return &macd{
		macd: ta.MACD(fastPeriod, slowPeriod, signalPeriod),
		idx:  (fastPeriod + slowPeriod + signalPeriod) / 3,
	}
}

func MACDWithResistance(resistance, fastPeriod, slowPeriod, signalPeriod int, fn ta.MovingAverageFunc) Strategy {
	if resistance < 0 {
		resistance = 0
	}
	m := ta.MACD(fastPeriod, slowPeriod, signalPeriod)
	if fn != nil {
		m = ta.MACDExt(fastPeriod, slowPeriod, signalPeriod, fn)
	}
	return &macd{
		macd: m,
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

func (s *macd) Setup(candles []*Candle) {
	for _, c := range candles {
		s.macd.Update(c.Close)
	}
}

func (s *macd) Update(c *Candle) (buy, sell bool) {
	v := s.macd.Update(c.Close)

	switch {
	case s.idx > 0:
		s.idx--
	case decimal.Crossunder(v, s.last, 0):
		s.dir = -1
	case decimal.Crosover(v, s.last, 0):
		s.dir = 1
	case v > s.last && s.dir > 0:
		s.dir++
	case v < s.last && s.dir < 0:
		s.dir--
	default:
		s.dir = 0
	}

	s.last = v
	return s.dir > s.res, s.dir < -s.res
}
