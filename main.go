package main

import (
	"time"
)

func parseTime(value string) time.Time {
	t, _ := time.Parse("2006-01-02", value)
	return t
}

func main() {
	/*symbol := "ETHUSDT"
	tUnitSec := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
	for ; tUnitSec < tUnitSec+(86400*365); tUnitSec += 86400 {
		fileName := "d:\\05\\market\\" + symbol + "\\" + time.Unix(tUnitSec, 0).Format("2006-01-02") + ".csv"
		client/LoadData(symbol, time.Unix(tUnitSec, 0), fileName)
	}*/
}
