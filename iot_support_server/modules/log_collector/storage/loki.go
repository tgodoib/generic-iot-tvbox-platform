package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log_collector/record"
	"net/http"
	"strconv"
	"time"
)

type LokiStorage struct {
	url    string
	client *http.Client
}

func NewLokiStorage(url string) (*LokiStorage, error) {
	return &LokiStorage{
		url:    url,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"`
}

type LokiPayload struct {
	Streams []LokiStream `json:"streams"`
}

func (s *LokiStorage) Write(record record.Record) error {

	// TODO: Check timestamp unit

	tags := map[string]string{}
	for k, v := range record.Tags {
		tags[k] = fmt.Sprint(v)
	}
	tags["deviceId"] = record.DeviceID
	tags["level"] = record.Level.String()

	payload := LokiPayload{
		Streams: []LokiStream{
			{
				Stream: tags,
				Values: [][2]string{
					{strconv.FormatInt(record.ReceivedAt*int64(time.Millisecond), 10), record.Msg},
				},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := s.client.Post(
		s.url+"/loki/api/v1/push",
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("loki error: %s", string(body))
	}

	return nil
}
