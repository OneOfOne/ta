package strategy_test

import (
	"fmt"
	"sync"
	"testing"

	"go.oneofone.dev/ta"
	"go.oneofone.dev/ta/csvticks"
	"go.oneofone.dev/ta/decimal"
	"go.oneofone.dev/ta/strategy"
)

func TestStrategy(t *testing.T) {
	const fname = "../testdata/AAPL.txt.gz"
	defer csvticks.SetDefaultTimeFormat(csvticks.SetDefaultTimeFormat("2006-01-02 15:04:05"))

	// load the data, the csv format is
	// timestamp, open, high, low, close, volume
	ticks, err := csvticks.Load(fname, csvticks.Mapping{
		Decoder: csvticks.GzipDecoder,

		FillSymbol: "AAPL",
		TS:         csvticks.CSVIndex(0),
		Open:       csvticks.CSVIndex(1),
		High:       csvticks.CSVIndex(2),
		Low:        csvticks.CSVIndex(3),
		Close:      csvticks.CSVIndex(4),
		Volume:     csvticks.CSVIndex(5),
	})
	if err != nil {
		panic(err)
	}

	acc := strategy.NewAccount("10F1", strategy.AccountOptions{
		BuyingPower:        2000,
		MaxSharesPerSymbol: 10,
		CanShort:           false,
		ReuseCash:          true,
	})

	// define a buy strategy, using either RSI or MACD
	buyStrat := strategy.Merge(strategy.RSI(26, 40, 80, nil), strategy.MACDWithResistance(5, 14, 26, 9, nil))

	// define a sell strategy, using macd
	// sellStrat := strategy.MACDWithResistance(10, 14, 26, 9)

	// we're using different strategies for buying and selling
	strat := strategy.Mixed(buyStrat, buyStrat)

	// apply the strategy, initial balance is 2000 dollars, maximum shares to hold at a time is 10
	res := strategy.ApplySlice(acc, strat, "AAPL", ticks)
	bp, hold, sv := acc.Balance()
	fmt.Printf("bought: %v, sold: %v, assets (%v): $%.2f, balance left: $%.2f ($%.2f on hold), total: $%.2f, profit/loss: $%.2f (%.2f%%)\n",
		res.Bought, res.Sold, res.Held, sv, bp, hold, res.Total(), res.PL(), res.PLPerc())
	if res.PLPerc() < 2 {
		t.Fatal("res.PLPerc() < 2")
	}
}

func TestSelectBestMACDStrat(t *testing.T) {
	ticks := loadSPY()
	var fidx, res, fast, slow, period int
	var pl decimal.Decimal
	var wg sync.WaitGroup
	var mux sync.Mutex
	for i, fn := range []ta.MovingAverageFunc{ta.DEMA, ta.EMA, ta.SMA, ta.TEMA, ta.WMA} {
		for f := 5; f < 41; f++ {
			for s := 20; s < 40; s++ {
				for p := 5; p < 30; p += 5 {
					for r := 5; r < 25; r += 5 {
						wg.Add(1)
						i, fn, f, s, p, r := i, fn, f, s, p, r
						go func() {
							defer wg.Done()
							acc := strategy.NewAccount("10F1", strategy.AccountOptions{
								BuyingPower:        2000,
								MaxSharesPerSymbol: 10,
								CanShort:           false,
								ReuseCash:          true,
							})
							_ = fn
							strat := strategy.MACDWithResistance(r, f, s, p, fn)
							// strat := strategy.MergeMatchAll(strategy.VWAP(5, -20), strategy.RSI(rsiP, int(slow), int(fast), fn))

							// apply the strategy, initial balance is 2000 dollars, maximum shares to hold at a time is 10
							tx := strategy.ApplySlice(acc, strat, "SPY", ticks)
							// t.Logf("With %v %v %v, got: %v", fast, slow, rsiP, res.PLPerc())
							mux.Lock()
							if tx.PLPerc() > pl {
								fast, slow, period, pl = f, s, p, tx.PLPerc()
								res, fidx = r, i
								t.Log(acc.Balance())
								t.Log("-", i, r, f, s, p, pl)
							}
							mux.Unlock()
						}()

					}
				}
			}
		}
	}
	wg.Wait()
	t.Log("best:", fidx, res, fast, slow, period, pl)
}

func loadSPY() csvticks.Ticks {
	const fname = "../testdata/SPY-20221001.csv.gz"
	defer csvticks.SetDefaultTimeFormat(csvticks.SetDefaultTimeFormat("2006-01-02 15:04:05"))

	// rsrc := rand.New(rand.NewSource(42))
	// load the data, the csv format is
	// timestamp, open, high, low, close, volume
	ticks, err := csvticks.Load(fname, csvticks.Mapping{
		SkipFirstRow: true,
		Decoder:      csvticks.GzipDecoder,

		FillSymbol: "SPY",
		TS:         csvticks.CSVIndex(0),
		Open:       csvticks.CSVIndex(1),
		High:       csvticks.CSVIndex(2),
		Low:        csvticks.CSVIndex(3),
		Close:      csvticks.CSVIndex(4),
		Volume:     csvticks.CSVIndex(6),
	})
	if err != nil {
		panic(err)
	}
	return ticks
}
