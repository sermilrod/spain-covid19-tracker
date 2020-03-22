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
	for _, data := range csvData[1 : len(csvData)-2] {
		ccaaCases := NewCCAACases(data)
		if err := col.UpdateCases(ccaaCases); err != nil {
			return err
		}
		if err := col.UpdateCasesPer100K(ccaaCases); err != nil {
			return err
		}
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
