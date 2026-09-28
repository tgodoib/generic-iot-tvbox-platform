package storage

import "log_collector/record"

type Storage interface {
	Write(entry record.Record) error
}
