package ta

import (
	"strconv"
)

func checkPeriod(period, min int) {
	if period < min {
		panic("period < " + strconv.Itoa(min))
	}
}
