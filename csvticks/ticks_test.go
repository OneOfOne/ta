package csvticks

import (
	"compress/gzip"
	"io"
	"testing"
)

func TestTicks(t *testing.T) {
	const fname = "testdata/AAPL.txt.gz"
	defer SetDefaultTimeFormat(SetDefaultTimeFormat("2006-01-02 15:04:05"))

	ticks, err := Load(fname, Mapping{
		Decoder: func(r io.Reader) (io.Reader, error) {
			return gzip.NewReader(r)
		},

		FillSymbol: "AAPL",
		TS:         CSVIndex(0),
		Open:       CSVIndex(1),
		High:       CSVIndex(2),
		Low:        CSVIndex(3),
		Close:      CSVIndex(4),
		Volume:     CSVIndex(5),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("ticks: %d", len(ticks))
}
