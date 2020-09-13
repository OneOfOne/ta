# go.oneofone.dev/ta [![Documentation](https://pkg.go.dev/badge/go.oneofone.dev/ta)](https://pkg.go.dev/go.oneofone.dev/ta@master?tab=doc#pkg-constants) [![Testing status](https://github.com/OneOfOne/ta/workflows/testing/badge.svg)](https://github.com/OneOfOne/ta/actions) [![Coverage Status](https://coveralls.io/repos/github/OneOfOne/ta/badge.svg?branch=master)](https://coveralls.io/github/OneOfOne/ta?branch=master) [![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause) [![Go Report Card](https://goreportcard.com/badge/go.oneofone.dev/ta)](https://goreportcard.com/report/go.oneofone.dev/ta) ![stability-experimental](https://img.shields.io/badge/stability-experimental-orange.svg)

A Go Technical Analysis library, mostly inspired by python's [TA-Lib](https://pypi.org/project/TA-Lib/) and the port by [markcheno](https://github.com/markcheno/go-talib).
It Can be used for backtesting or eventually creating strategies for live trading.

## Features

* Tries to be compatible with the python version for testing, however all the functions supports partial updates to help working with live data.
* Going for a healthy mix of speed and accuracy.
* Includes option related functions.

## Install

```bash
go get -u go.oneofone.dev/ta
```

## Status: **PRE ALPHA**

* the API is not stable at all
* Missing a lot of indicators compared to the python version or markcheno's port

## TODO

* Port more functions
* More testing / benchmarks
* Stablize the API
* Documentation

## Example

```go
package main

import (
	"fmt"
	"github.com/markcheno/go-quote"
	"go.oneofone.dev/ta"
)

func main() {
	spy, _ := quote.NewQuoteFromYahoo("spy", "2016-01-01", "2016-04-01", quote.Daily, true)
	fmt.Print(spy.CSV())
	dema, _ := ta.New(spy.Close).DEMA(10)
	fmt.Println(dema)
}
```

* all the `*_test.go` files should have more examples.

## Dependencies

* [gonum/gonum](https://github.com/gonum/gonum) (*BSD-3-Clause*)

## References

Without those libraries and their documentation, this wouldn't have been possible.

* [mrjbq7/ta-lib](https://github.com/mrjbq7/ta-lib) (*MIT?*)
* [markcheno/go-talib](https://github.com/markcheno/go-talib) (M*IT*)
* [TulipCharts/tulipindicators](https://github.com/TulipCharts/tulipindicators) (*LGPL v3.0*)
* [greyblake/ta-rs](https://github.com/greyblake/ta-rs) (*MIT*)
* [MattL922/black-scholes](https://github.com/MattL922/black-scholes) (*MIT*)
* [MattL922/implied-volatility](https://github.com/MattL922/implied-volatility) (*MIT*)

## License

[BSD-3-Clause](LICENSE)
