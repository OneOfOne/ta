package csvticks

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"strconv"

	"go.oneofone.dev/ta"
	"go.oneofone.dev/ta/decimal"
)

type Idx int

func (i *Idx) val() int {
	if i != nil {
		return int(*i)
	}
	return -1
}

// CSVIndex helps define the index of the field in a csv file
func CSVIndex(i int) *Idx {
	return (*Idx)(&i)
}

// Mapping defines how to read a CSV file
type Mapping struct {
	SkipFirstRow bool
	Decoder      func(r io.Reader) (io.Reader, error)

	TS     *Idx
	Symbol *Idx
	Open   *Idx
	High   *Idx
	Low    *Idx
	Close  *Idx
	Volume *Idx

	FillSymbol string

	maxIndex int
}

func (m *Mapping) init() error {
	m.maxIndex = decimal.MaxInt(m.TS.val(), m.Symbol.val(), m.Open.val(), m.High.val(), m.Low.val(), m.Close.val(), m.Volume.val())
	if m.maxIndex == -1 {
		return ErrMissingMapping
	}
	return nil
}

func (m *Mapping) get(row []string) (_ *Tick, err error) {
	if m.maxIndex >= len(row) {
		log.Printf("bad row? %v", row)
		return nil, err
	}

	var t Tick
	if v := m.TS.val(); v > -1 {
		t.TS = DateTime(row[v])
	}

	if v := m.Symbol.val(); v > -1 {
		t.Symbol = row[v]
	}

	if m.FillSymbol != "" {
		t.Symbol = m.FillSymbol
	}

	if v := m.Open.val(); v > -1 {
		if t.Open, err = decimal.ParseString(row[v]); err != nil {
			return
		}
	}

	if v := m.High.val(); v > -1 {
		if t.High, err = decimal.ParseString(row[v]); err != nil {
			return
		}
	}

	if v := m.Low.val(); v > -1 {
		if t.Low, err = decimal.ParseString(row[v]); err != nil {
			return
		}
	}

	if v := m.Close.val(); v > -1 {
		if t.Close, err = decimal.ParseString(row[v]); err != nil {
			return
		}
	}

	if v := m.Volume.val(); v > -1 {
		if t.Volume, err = strconv.ParseInt(row[v], 10, 64); err != nil {
			return
		}
	}

	return &t, nil
}

func Load(fname string, mapping Mapping) (Ticks, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadReader(f, mapping)
}

func LoadReader(r io.Reader, mapping Mapping) (tks Ticks, err error) {
	if err = mapping.init(); err != nil {
		return
	}

	if mapping.Decoder != nil {
		if r, err = mapping.Decoder(r); err != nil {
			return
		}
	}

	var (
		cf  = csv.NewReader(r)
		rec []string
		t   *Tick
	)

	cf.ReuseRecord = true

	if mapping.SkipFirstRow {
		if _, err = cf.Read(); err != nil {
			return
		}
	}

	for {
		if rec, err = cf.Read(); err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return
		}

		if t, err = mapping.get(rec); err != nil {
			return
		}

		tks = append(tks, t)
	}
}

// Tick represents a single tick
type Tick struct {
	TS     DateTime   `json:"date"`
	Symbol string     `json:"symbol"`
	Open   ta.Decimal `json:"open"`
	High   ta.Decimal `json:"high"`
	Low    ta.Decimal `json:"low"`
	Close  ta.Decimal `json:"close"`
	Volume int64      `json:"volume"`
}

type Ticks []*Tick

func (tks Ticks) Filter(fn func(t *Tick) bool, inPlace bool) Ticks {
	var out Ticks
	if inPlace {
		out = tks[:0]
	}
	for _, t := range tks {
		if fn(t) {
			out = append(out, t)
		}
	}
	return out[:len(out):len(out)]
}

// BySymbol returns ticks with only the given symbol
func (tks Ticks) BySymbol(symbol string) Ticks {
	return tks.Filter(func(t *Tick) bool { return t.Symbol == symbol }, false)
}

// Open returns only Open values as a TA
func (tks Ticks) Open() *ta.TA {
	out := ta.NewSize(len(tks), true)
	for _, t := range tks {
		out.Append(t.Open)
	}
	return out
}

// High returns only High values as a TA
func (tks Ticks) High() *ta.TA {
	out := ta.NewSize(len(tks), true)
	for _, t := range tks {
		out.Append(t.High)
	}
	return out
}

// Low returns only Low values as a TA
func (tks Ticks) Low() *ta.TA {
	out := ta.NewSize(len(tks), true)
	for _, t := range tks {
		out.Append(t.Low)
	}
	return out
}

// Close returns only Close values as a TA
func (tks Ticks) Close() *ta.TA {
	out := ta.NewSize(len(tks), true)
	for _, t := range tks {
		out.Append(t.Close)
	}
	return out
}
