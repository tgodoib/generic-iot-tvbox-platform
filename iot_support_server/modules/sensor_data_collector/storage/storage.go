package storage

import (
	"context"
	"sensor_data_collector/record"
)

type Storage interface {
	Write(ctx context.Context, entry record.SensorDataRecord) error
}
