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

func (m *merge) Update(t *Tick) (buy, sell bool) {
	buys, sells := 0, 0
	for _, s := range m.strats {
		buy, sell = s.Update(t)
		if buy {
			buys++
		}
		if sell {
			sells++
		}
	}
	if m.matchAny {
		return buys > 0, sells > 0
	}
	return buys == len(m.strats), sells == len(m.strats)
}

func Mixed(buyStrat, sellStrat Strategy) Strategy {
	return &mixed{buy: buyStrat, sell: sellStrat}
}

type mixed struct {
	buy, sell Strategy
}

func (m *mixed) Update(t *Tick) (buy, sell bool) {
	buy, _ = m.buy.Update(t)
	_, sell = m.sell.Update(t)
	return
}
