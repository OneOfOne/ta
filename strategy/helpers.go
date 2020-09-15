package strategy

func Merge(strats ...Strategy) Strategy {
	if len(strats) < 2 {
		return strats[0]
	}
	return &merge{strats: strats, matchAny: true}
}

func MergeMatchAll(strats ...Strategy) Strategy {
	if len(strats) < 2 {
		return strats[0]
	}
	return &merge{strats: strats, matchAny: false}
}

type merge struct {
	strats   []Strategy
	matchAny bool
}

func (s *merge) Setup(candles []*Candle) {
	for _, s := range s.strats {
		s.Setup(candles)
	}
}

func (s *merge) Update(t *Candle) (buy, sell bool) {
	buys, sells := 0, 0
	for _, st := range s.strats {
		buy, sell = st.Update(t)
		if buy {
			buys++
		}
		if sell {
			sells++
		}
	}
	if s.matchAny {
		return buys > 0, sells > 0
	}
	return buys == len(s.strats), sells == len(s.strats)
}

func Mixed(buyStrat, sellStrat Strategy) Strategy {
	return &mixed{buy: buyStrat, sell: sellStrat}
}

type mixed struct {
	buy, sell Strategy
}

func (s *mixed) Setup(candles []*Candle) {
	s.buy.Setup(candles)
	s.sell.Setup(candles)
}

func (s *mixed) Update(t *Candle) (buy, sell bool) {
	buy, _ = s.buy.Update(t)
	_, sell = s.sell.Update(t)
	return
}
