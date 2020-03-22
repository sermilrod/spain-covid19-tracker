package collector

import (
	"strconv"
	"time"
)

func NewCCAACases(data []string) *CCAACases {
	code := data[0]

	// NOTE: ignoring errors will set default value as 0
	//	     this is required when original source data is empty
	date, _ := time.Parse("02/01/2006", data[1])
	cases, _ := strconv.ParseInt(data[2], 10, 64)
	hospitalised, _ := strconv.ParseInt(data[3], 10, 64)
	critical, _ := strconv.ParseInt(data[4], 10, 64)
	deaths, _ := strconv.ParseInt(data[5], 10, 64)

	return &CCAACases{
		code:         code,
		date:         date,
		cases:        cases,
		hospitalised: hospitalised,
		critical:     critical,
		deaths:       deaths,
	}
}