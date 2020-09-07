# ta

[![PkgGoDev](https://pkg.go.dev/badge/go.oneofone.dev/ta@master)](https://pkg.go.dev/go.oneofone.dev/ta@master?tab=doc#pkg-constants)
[![testing](https://github.com/OneOfOne/ta/workflows/testing/badge.svg)](https://github.com/OneOfOne/ta/actions)
[![Coverage Status](https://coveralls.io/repos/github/OneOfOne/ta/badge.svg?branch=master)](https://coveralls.io/github/OneOfOne/ta?branch=master)

A Go Technical Analysis library, mostly inspired by python's [TA-Lib](https://pypi.org/project/TA-Lib/) and the port by [markcheno](https://github.com/markcheno/go-talib).

## Features

* Tries to be compatible with the python version for testing, however all functions support partial updates.
* Going for a healthy mix of speed and accuracy.

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

## References

* [mrjbq7/ta-lib](https://github.com/mrjbq7/ta-lib) (MIT?)
* [markcheno/go-talib](https://github.com/markcheno/go-talib) (MIT)
* [TulipCharts/tulipindicators](https://github.com/TulipCharts/tulipindicators) (LGPL v3.0)
* [greyblake/ta-rs](https://github.com/greyblake/ta-rs) (MIT)

## License

[BSD License](LICENSE)
