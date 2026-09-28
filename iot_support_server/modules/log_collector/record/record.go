package record

import (
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/fxamacker/cbor/v2"
)

type Record struct {
	DeviceID   string
	ReceivedAt int64          // milliseconds; set on MQTT arrival, not in CBOR
	Level      Level          `cbor:"level"`
	Timestamp  int64          `cbor:"timestamp,omitempty"`
	Tags       map[string]any `cbor:"tags,omitempty"`
	Msg        string         `cbor:"msg,omitempty"`
}

type Level int

const (
	LevelInfo Level = iota
	LevelDebug
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBUG"
	case LevelError:
		return "ERROR"
	}
	return "OTHER"
}

func NewRecord(rawMsg mqtt.Message) (Record, error) {
	var record Record

	err := cbor.Unmarshal(rawMsg.Payload(), &record)
	if err != nil {
		return Record{}, err
	}
	record.DeviceID = strings.Split(rawMsg.Topic(), "/")[2]
	record.ReceivedAt = time.Now().UnixMilli()

	if record.Timestamp == 0 {
		record.Timestamp = time.Now().Unix()
	}

	return record, nil
}
