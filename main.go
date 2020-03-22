package main

import (
	"log"

	"github.com/sermilrod/spain-covid19-tracker/pkg/collector"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	seriesURL          = kingpin.Flag("series-url", "Source of time series URL.").Default("https://covid19.isciii.es/resources/serie_historica_acumulados.csv").String()
	influxDBHost       = kingpin.Flag("influxdb-host", "InfluxDB Host.").Default("http://localhost").String()
	influxDBPort       = kingpin.Flag("influxdb-port", "InfluxDB Port.").Default("8086").String()
	influxDBDatabase   = kingpin.Flag("influxdb-database", "InfluxDB Database.").Default("spain_covid19").String()
	dbReadyTimeout     = kingpin.Flag("dbready-timeout", "Time to wait for DB readiness.").Default("120s").Duration()
	dbPingReadyTimeout = kingpin.Flag("dbready-ping-timeout", "Time to wait for DB ping probes.").Default("5s").Duration()
	dbReadyTick        = kingpin.Flag("dbready-tick", "Tick for DB ping probes.").Default("1s").Duration()
)

func init() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
}

func main() {
	c := collector.NewCovid19Collector(
		*influxDBHost,
		*influxDBPort,
		*influxDBDatabase,
		*seriesURL,
	)
	/*if ok, err := c.WaitForDBReady(*dbReadyTick, *dbReadyTimeout, *dbPingReadyTimeout); err != nil || !ok {
		log.Fatalln(err)
	}*/
	if err := c.UpdateStatus(); err != nil {
		log.Fatalln(err)
	}
}
