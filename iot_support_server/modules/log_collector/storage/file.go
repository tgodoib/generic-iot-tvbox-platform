package storage

import (
	"encoding/json"
	"log_collector/record"
	"os"
)

type FileStorage struct {
	file *os.File
	enc  *json.Encoder
}

func NewFileStorage(path string) (*FileStorage, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		file: f,
		enc:  json.NewEncoder(f),
	}, nil
}

func (s *FileStorage) Write(entry record.Record) error {
	if err := s.enc.Encode(entry); err != nil {
		return err
	}
	return s.file.Sync()
}
