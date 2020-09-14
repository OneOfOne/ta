package strategy_test

import (
	"fmt"
	"testing"

	"go.oneofone.dev/ta/csvticks"
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

	// we only need the close data
	closes := ticks.Close()

	// define a buy strategy, using either RSI or MACD
	buyStrat := strategy.Merge(strategy.RSI(26, 40, 80), strategy.MACDWithResistance(5, 14, 26, 9))

	// define a sell strategy, using macd
	sellStrat := strategy.MACDWithResistance(10, 14, 26, 9)

	// we're using different strategies for buying and selling
	strat := strategy.Mixed(buyStrat, sellStrat)

	// apply the strategy, initial balance is 2000 dollars, maximum shares to hold at a time is 10
	res := strategy.ApplySlice(acc, strat, "AAPL", closes)
	bp, hold, sv := acc.Balance()
	fmt.Printf("bought: %v, sold: %v, assets (%v): $%.2f, balance left: $%.2f ($%.2f on hold), total: $%.2f, profit/loss: $%.2f (%.2f%%)\n",
		res.Bought, res.Sold, res.Held, sv, bp, hold, res.Total(), res.PL(), res.PLPerc())
	if res.PLPerc() < 2 {
		t.Fatal("res.PLPerc() < 2")
	}
}
