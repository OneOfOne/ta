package ta

type AggFunc = func(*TA) Decimal

type Stream interface {
	Chan() <-chan Decimal
	Update(v Decimal)
	Close()
}

func StreamFromStudy(s Study, blockOnFull bool) Stream {
	return &studyStream{
		s:     s,
		ch:    make(chan Decimal, s.Len()),
		block: blockOnFull,
	}
}

type studyStream struct {
	s  Study
	ch chan Decimal

	block bool
}

func (a *studyStream) Update(v Decimal) {
	v = a.s.Update(v)
	if a.block {
		a.ch <- v
	} else {
		select {
		case a.ch <- v:
		default:
		}
	}
}

func (a *studyStream) Chan() <-chan Decimal {
	return a.ch
}

func (a *studyStream) Close() {
	close(a.ch)
}

func Aggregate(period int, blockOnFull bool) Stream {
	return AggregateFn((*TA).Avg, period, blockOnFull)
}

func AggregateFn(fn AggFunc, period int, blockOnFull bool) Stream {
	return &agg{
		d:     NewSize(period, true),
		ch:    make(chan Decimal, period),
		block: blockOnFull,
	}
}

type agg struct {
	d     *TA
	fn    AggFunc
	ch    chan Decimal
	block bool
}

func (a *agg) Setup(d *TA) {
	for i := 0; i < d.Len(); i++ {
		a.Update(d.Get(i))
	}
}

func (a *agg) Update(v Decimal) {
	if a.d.Append(v).Len() == a.d.Cap() {
		v = a.fn(a.d)
		if a.block {
			a.ch <- v
		} else {
			select {
			case a.ch <- v:
			default:
			}
		}
		a.d.Trunc(0)
	}
}

func (a *agg) Chan() <-chan Decimal {
	return a.ch
}

func (a *agg) Close() {
	close(a.ch)
}
