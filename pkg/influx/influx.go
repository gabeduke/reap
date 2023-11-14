package influx

import (
	"context"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type Influx struct {
	database     string
	influxClient influxdb2.Client
	logger       *log.Logger
}

func NewClient(url string) (*Influx, error) {
	token := os.Getenv("INFLUXDB_TOKEN")

	influx := &Influx{
		database: url,
		logger:   log.New(),
	}
	influx.influxClient = influxdb2.NewClient(url, token)

	return influx, nil
}

// WritePoint writes a point to the database
func (i *Influx) WritePoint(sensor, source string, value float64, tags map[string]string) error {
	org := "leetserve"
	bucket := source
	writeAPI := i.influxClient.WriteAPIBlocking(org, bucket)

	fields := map[string]interface{}{
		"value": value,
	}

	point := write.NewPoint(sensor, tags, fields, time.Now())

	return writeAPI.WritePoint(context.Background(), point)
}
