package collector

import (
	"net/http"
	"time"
)

type Covid19Collector struct {
	seriesURL, database, dbAddr string
	httpClient                  *http.Client
}

type CCAACases struct {
	date                                  time.Time
	cases, hospitalised, critical, deaths int64
	code                                  string
}
