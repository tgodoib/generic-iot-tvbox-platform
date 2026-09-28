package storage

import (
	"context"
	"sensor_data_collector/record"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type InfluxV2Storage struct {
	client influxdb2.Client
	api    api.WriteAPIBlocking
}

func NewInfluxV2Storage(url, token, org, bucket string) (*InfluxV2Storage, error) {
	client := influxdb2.NewClient(url, token)
	writeApi := client.WriteAPIBlocking(org, bucket)

	return &InfluxV2Storage{client: client, api: writeApi}, nil
}

func (i InfluxV2Storage) Write(ctx context.Context, entry record.SensorDataRecord) error {
	p := influxdb2.
		NewPointWithMeasurement(entry.Measurement).
		AddTag("device_id", entry.DeviceID).
		AddTag("sensor_id", entry.SensorID).
		AddField("value", entry.Value).
		SetTime(time.Unix(entry.Timestamp, 0))

	return i.api.WritePoint(ctx, p)
}
