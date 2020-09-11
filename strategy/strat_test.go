package strategy_test

import (
	"fmt"

	"go.oneofone.dev/ta/csvticks"
	"go.oneofone.dev/ta/strategy"
)

func ExampleStrategy() {
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

	closes := ticks.Close() // we only need the close data

	// define a buy strategy, using either RSI or MACD
	buyStrat := strategy.Merge(false, strategy.RSI(26, 40, 80), strategy.MACDWithResistance(10, 14, 26, 9))

	// define a sell strategy, using macd
	sellStrat := strategy.MACDWithResistance(10, 14, 26, 9)
	strat := strategy.Mixed(buyStrat, sellStrat)
	// apply the strategy, initial balance is 2000 dollars, maximum shares to hold at a time is 10
	res := strategy.Apply(strat, "AAPL", closes, 25000, 50)

	fmt.Printf("bought: %v, sold: %v, assets (%v): $%.3f, balance left: $%.3f, total: $%.3f, gain/loss: $%.2f (%.2f%%)\n",
		res.Bought, res.Sold, res.NumShares(), res.SharesValue(), res.Balance, res.Total(), res.PL(), res.PLPerc())

	// Output:
	// bought: 300, sold: 250, assets (50): $15773.500, balance left: $10245.355, total: $26018.855, gain/loss: $1018.86 (3.91%)
}
