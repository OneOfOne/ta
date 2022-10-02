package strategy

import (
	"go.oneofone.dev/ta"
	"go.oneofone.dev/ta/decimal"
)

func VWAP(up, down Decimal) Strategy {
	return &vwap{
		vwap: ta.VWAPBands(up, down),
		idx:  int(decimal.Max(up, down)),
	}
}

type vwap struct {
	vwap   ta.MultiVarStudy
	up, dn Decimal
	idx    int
	dir    int8
}

func (s *vwap) Setup(candles []*Candle) {
	for _, c := range candles {
		s.vwap.Update(c.Close)
	}
}

func (s *vwap) Update(c *Candle) (buy, sell bool) {
	v := s.vwap.UpdateAll(Decimal(c.Volume), c.Close)
	switch {
	case s.idx > 0:
		s.idx--
	case decimal.Crosover(v[0], s.up, v[1]):
		s.dir = 1
	case decimal.Crossunder(v[0], s.dn, v[2]):
		s.dir = -1
	default:
		s.dir = 0
	}
	s.up, s.dn = v[1], v[2]

	return s.dir > 0, s.dir < 0
}
