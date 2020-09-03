package decimal

import (
	"testing"
)

func TestDecimal_Add(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]Decimal{
		{"2", "3"}:                     5,
		{"2454495034", "3451204593"}:   5905699627,
		{"24544.95034", ".3451204593"}: 24545.2954604593,
		{".1", ".1"}:                   0.2,
		{".1", "-.1"}:                  0,
		{"0", "1.001"}:                 1.001,
	}

	for inp, res := range inputs {
		a, err := ParseString(inp.a)
		if err != nil {
			t.FailNow()
		}
		b, err := ParseString(inp.b)
		if err != nil {
			t.FailNow()
		}
		c := a.Add(b)
		if !c.Equal(res) {
			t.Errorf("expected %s, got %s", res, c.String())
		}
	}
}
