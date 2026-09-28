package storage

import (
	"context"
	"health_checker/record"
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

func (i InfluxV2Storage) Write(ctx context.Context, entry record.StateRecord) error {
	p := influxdb2.
		NewPointWithMeasurement("state").
		AddTag("device_id", entry.DeviceID).
		SetTime(time.Unix(entry.Timestamp, 0))

	for k := range entry.States {
		p.AddField(k, entry.States[k])
	}

	//log.Println("write", write.PointToLineProtocol(p, time.Second))

	return i.api.WritePoint(ctx, p)
}
