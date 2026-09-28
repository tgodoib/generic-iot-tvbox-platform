package storage

import (
	"context"
	"health_checker/record"
)

type Storage interface {
	Write(ctx context.Context, entry record.StateRecord) error
}
