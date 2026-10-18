package reap

import (
	"log/slog"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// Influx writes points with the client's batching, non-blocking write API.
type Influx struct {
	client influxdb2.Client
	api    api.WriteAPI
}

// NewInflux returns a Writer for org/bucket. Write errors are logged, not returned.
func NewInflux(url, token, org, bucket string, log *slog.Logger) *Influx {
	c := influxdb2.NewClient(url, token)
	w := c.WriteAPI(org, bucket)
	go func() {
		for err := range w.Errors() {
			log.Error("influx write failed", "bucket", bucket, "err", err)
		}
	}()
	return &Influx{client: c, api: w}
}

// Write queues a point with the current time.
func (i *Influx) Write(p Point) {
	i.api.WritePoint(influxdb2.NewPoint(p.Measurement, p.Tags, map[string]interface{}{"value": p.Value}, time.Now()))
}

// Close flushes pending points and closes the client.
func (i *Influx) Close() {
	i.api.Flush()
	i.client.Close()
}
