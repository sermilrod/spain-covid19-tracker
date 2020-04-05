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
	date         time.Time
	cases        int64
	hospitalised int64
	critical     int64
	deaths       int64
	recovered    int64
	active       int64
	code         string
}
