package client

import (
	_ "github.com/influxdata/influxdb1-client"
	influxdbclient "github.com/influxdata/influxdb1-client/v2"
)

func NewInfluxDBClient(addr string) (influxdbclient.Client, error) {
	c, err := influxdbclient.NewHTTPClient(influxdbclient.HTTPConfig{
		Addr: addr,
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}
