package influx

import (
	"context"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/user/can-server/config"
)

type Writer struct {
	enabled     bool
	org         string
	bucket      string
	measurement string
	client      influxdb2.Client
	writeAPI    api.WriteAPIBlocking
}

type MQTTRecord struct {
	Topic       string
	HostCode    string
	ChannelCode string
	Payload     string
	Value       *float64
	StrValue    *string
	BoolValue   *bool
	Quality     int
	ReceivedAt  time.Time
}

func NewWriter(cfg config.InfluxDBConfig) *Writer {
	if cfg.Token == "" {
		log.Println("influxdb token empty, mqtt raw data will not be written to influxdb")
		return &Writer{}
	}
	client := influxdb2.NewClient(cfg.URL, cfg.Token)
	return &Writer{
		enabled:     true,
		org:         cfg.Org,
		bucket:      cfg.Bucket,
		measurement: cfg.Measurement,
		client:      client,
		writeAPI:    client.WriteAPIBlocking(cfg.Org, cfg.Bucket),
	}
}

func (w *Writer) Close() {
	if w.client != nil {
		w.client.Close()
	}
}

func (w *Writer) WriteMQTT(ctx context.Context, record MQTTRecord) error {
	if !w.enabled {
		return nil
	}
	if record.ReceivedAt.IsZero() {
		record.ReceivedAt = time.Now()
	}
	fields := map[string]any{
		"payload": record.Payload,
		"quality": record.Quality,
	}
	if record.Value != nil {
		fields["value"] = *record.Value
	}
	if record.StrValue != nil {
		fields["str_value"] = *record.StrValue
	}
	if record.BoolValue != nil {
		fields["bool_value"] = *record.BoolValue
	}

	tags := map[string]string{
		"source":       "mqtt",
		"topic":        record.Topic,
		"host_code":    record.HostCode,
		"channel_code": record.ChannelCode,
	}
	point := write.NewPoint(w.measurement, tags, fields, record.ReceivedAt)
	return w.writeAPI.WritePoint(ctx, point)
}
