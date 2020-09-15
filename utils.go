package ta

import (
	"strconv"
	"time"

	"go.oneofone.dev/ta/decimal"
)

func checkPeriod(period, min int) {
	if period < min {
		panic("period < " + strconv.Itoa(min))
	}
}

func AggPipe(aggPeriod time.Duration, in <-chan Decimal) <-chan Decimal {
	return decimal.AggPipe(aggPeriod, in)
}
