package csvticks

import (
	"compress/gzip"
	"io"
	"testing"
)

func TestTicks(t *testing.T) {
	const fname = "/tmp/aapl.csv.gz"
	defer SetDefaultTimeFormat(SetDefaultTimeFormat("2006-01-02 15:04:05"))

	ticks, err := Load(fname, Mapping{
		Decoder: func(r io.Reader) (io.Reader, error) {
			return gzip.NewReader(r)
		},

		FillSymbol: "AAPL",
		TS:         FieldIndex(0),
		Open:       FieldIndex(1),
		High:       FieldIndex(2),
		Low:        FieldIndex(3),
		Close:      FieldIndex(4),
		Volume:     FieldIndex(5),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("ticks: %d", len(ticks))
}
