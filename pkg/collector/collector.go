package collector

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/influxdata/influxdb1-client"
	influxdbclient "github.com/influxdata/influxdb1-client/v2"
	"github.com/sermilrod/spain-covid19-tracker/pkg/client"
	"github.com/sermilrod/spain-covid19-tracker/pkg/parser"
)

func NewCovid19Collector(addr, port, database, seriesURL string) *Covid19Collector {
	var b bytes.Buffer

	b.WriteString(addr)
	b.WriteString(":")
	b.WriteString(port)

	return &Covid19Collector{
		seriesURL: seriesURL,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		database: database,
		dbAddr:   b.String(),
	}
}

func (col *Covid19Collector) WriteMetricWithTimestamp(
	c influxdbclient.Client,
	measurement string,
	tags map[string]string,
	fields map[string]interface{},
	ts time.Time) error {
	defer c.Close()

	// Create a new point batch
	bp, _ := influxdbclient.NewBatchPoints(influxdbclient.BatchPointsConfig{
		Precision: "s",
		Database:  col.database,
	})

	pt, err := influxdbclient.NewPoint(measurement, tags, fields, ts)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// Write the batch
	return c.Write(bp)
}

func (col *Covid19Collector) UpdateStatus() error {
	totalCases := make(map[time.Time]map[string]int64)
	activesYesterday := make(map[string]int64)
	recoveredYesterday := make(map[string]int64)
	log.Println("Fetching data")
	dataRes, err := col.httpClient.Get(col.seriesURL)
	if err != nil {
		return err
	}
	csvData, err := parser.ParseCSV(dataRes.Body, -1)
	if err != nil {
		return err
	}
	log.Println("Persisting data")
	info := csvData[1 : len(csvData)-2]
	for index, data := range info {
		ccaaCases, casesErr := NewCCAACases(data)
		if casesErr != nil {
			log.Println(casesErr)
			continue
		}
		if err := col.UpdateCases(ccaaCases); err != nil {
			return err
		}
		if err := col.UpdateCasesPer100K(ccaaCases); err != nil {
			return err
		}
		if index > 0 {
			if ccaaCases.recovered != 0 && recoveredYesterday[ccaaCases.code] != 0 {
				infectionRate := float64(ccaaCases.active) / float64(activesYesterday[ccaaCases.code])
				recoveryRate := float64(ccaaCases.recovered) / float64(recoveredYesterday[ccaaCases.code])
				if err := col.UpdateRate(
					"reproductive_rate",
					infectionRate,
					recoveryRate,
					ccaaISOCode[ccaaCases.code],
					ccaaCases.date); err != nil {
					return err
				}
			}
		}
		activesYesterday[ccaaCases.code] = ccaaCases.active
		recoveredYesterday[ccaaCases.code] = ccaaCases.recovered
		inner, ok := totalCases[ccaaCases.date]
		if !ok {
			inner = make(map[string]int64)
			totalCases[ccaaCases.date] = inner
		}
		totalCases[ccaaCases.date]["cases"] += ccaaCases.cases
		totalCases[ccaaCases.date]["hospitalised"] += ccaaCases.hospitalised
		totalCases[ccaaCases.date]["critical"] += ccaaCases.critical
		totalCases[ccaaCases.date]["deaths"] += ccaaCases.deaths
		totalCases[ccaaCases.date]["recovered"] += ccaaCases.recovered
		totalCases[ccaaCases.date]["active"] += ccaaCases.active
	}
	return col.UpdateTotalCases(totalCases)
}

func (col *Covid19Collector) UpdateTotalCases(tc map[time.Time]map[string]int64) error {
	var activesYesterday, recoveredYesterday, index int64
	tags := map[string]string{}

	for date, data := range tc {
		fields := map[string]interface{}{
			"cases":        data["cases"],
			"hospitalised": data["hospitalised"],
			"critical":     data["critical"],
			"deaths":       data["deaths"],
			"recovered":    data["recovered"],
			"active":       data["active"],
		}
		cc, err := client.NewInfluxDBClient(col.dbAddr)
		if err != nil {
			return err
		}
		if err := col.WriteMetricWithTimestamp(
			cc,
			fmt.Sprintf("%s_cases_total", col.database),
			tags,
			fields,
			date,
		); err != nil {
			return err
		}
		if index > 0 {
			if data["recovered"] != 0 && recoveredYesterday != 0 {
				infectionRate := float64(data["active"]) / float64(activesYesterday)
				recoveryRate := float64(data["recovered"]) / float64(recoveredYesterday)
				if err := col.UpdateRate(
					"reproductive_rate",
					infectionRate,
					recoveryRate,
					"total",
					date); err != nil {
					return err
				}
			}
		}
		activesYesterday = data["active"]
		recoveredYesterday = data["recovered"]
		index += 1
	}

	return nil
}

func (col *Covid19Collector) UpdateCases(c *CCAACases) error {
	// Create ccaa points and add them to batch
	tags := map[string]string{}
	fields := map[string]interface{}{
		"cases":        c.cases,
		"hospitalised": c.hospitalised,
		"critical":     c.critical,
		"deaths":       c.deaths,
		"recovered":    c.recovered,
		"active":       c.active,
	}
	cc, err := client.NewInfluxDBClient(col.dbAddr)
	if err != nil {
		return err
	}
	return col.WriteMetricWithTimestamp(
		cc,
		fmt.Sprintf("%s_cases_%s", col.database, ccaaISOCode[c.code]),
		tags,
		fields,
		c.date,
	)
}

func (col *Covid19Collector) UpdateCasesPer100K(c *CCAACases) error {
	tags := map[string]string{}
	casesPer100K := (float64(c.cases) / float64(ccaaInhabitatns[c.code])) * float64(100000)
	fields := map[string]interface{}{"cases": casesPer100K}

	cc, err := client.NewInfluxDBClient(col.dbAddr)
	if err != nil {
		return err
	}
	return col.WriteMetricWithTimestamp(
		cc,
		fmt.Sprintf("%s_cases_per_100000_%s", col.database, ccaaISOCode[c.code]),
		tags,
		fields,
		c.date,
	)
}

func (col *Covid19Collector) UpdateRate(metric string,
	rateToday, rateYesterday float64, code string,
	date time.Time) error {

	// Skip NaN when there is no data
	if rateYesterday < 1 {
		return nil
	}
	tags := map[string]string{}
	rate := rateToday / rateYesterday
	fields := map[string]interface{}{"rate": rate}

	cc, err := client.NewInfluxDBClient(col.dbAddr)
	if err != nil {
		return err
	}
	return col.WriteMetricWithTimestamp(
		cc,
		fmt.Sprintf("%s_%s_%s", col.database, metric, code),
		tags,
		fields,
		date,
	)
}

func (col *Covid19Collector) WaitForDBReady(tick, timeout, pTimeout time.Duration) (bool, error) {
	log.Println("Waiting for DB ready...")
	for {
		select {
		case <-time.After(timeout):
			return false, errors.New("timed out waiting for DB ready")
		case <-time.After(tick):
			c, err := client.NewInfluxDBClient(col.dbAddr)
			if err != nil {
				return false, nil
			} else {
				q := influxdbclient.NewQuery("show measurements", col.database, "")
				if response, err := c.Query(q); err == nil && response.Error() == nil {
					log.Printf("DB not ready: %s, trying again...", response.Results)
					return false, nil
				}
				return true, nil
			}
		}
	}
}
