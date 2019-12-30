package main

import (
	"context"
	influxdb "github.com/influxdata/influxdb-client-go"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var influx *influxdb.Client

func influxConnect(influxAddr string, token string) {
	var err error
	httpClient := &http.Client{}

	influx, err = influxdb.New(influxAddr, token, influxdb.WithHTTPClient(httpClient))
	if err != nil {
		log.Fatal(err)
	}
}

func pushInfluxDist(dist float32, ax uint16) {
	myMetrics := []influxdb.Metric{
		influxdb.NewRowMetric(
			map[string]interface{}{"level": dist,
				"ax": ax},
			"pegel-graefte",
			map[string]string{},
			time.Now()),
	}

	if _, err := influx.Write(context.Background(), "my-awesome-bucket", "my-very-awesome-org", myMetrics...); err != nil {
		log.Fatal(err) // as above use your own error handling here.
	}
}
