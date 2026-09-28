package record

import (
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/fxamacker/cbor/v2"
)

type MonitorRecord struct {
	DeviceID string
	Enabled  bool  `cbor:"enabled"`
	Timeout  int64 `cbor:"timeout"`
}

type StateRecord struct {
	DeviceID  string
	Timestamp int64          `cbor:"timestamp,omitempty"`
	States    map[string]any `cbor:"-"`
}

func NewMonitorRecord(m MQTT.Message) (MonitorRecord, error) {
	var record MonitorRecord

	err := cbor.Unmarshal(m.Payload(), &record)
	if err != nil {
		return MonitorRecord{}, err
	}

	record.DeviceID = strings.Split(m.Topic(), "/")[2]

	return record, nil
}

func NewStateRecord(m MQTT.Message) (StateRecord, error) {
	var record StateRecord

	var raw map[string]any
	err := cbor.Unmarshal(m.Payload(), &raw)
	if err != nil {
		return StateRecord{}, err
	}

	if ts, ok := raw["timestamp"].(int64); ok {
		record.Timestamp = ts
		delete(raw, "timestamp")
	} else {
		record.Timestamp = time.Now().Unix()
	}

	record.States = raw
	record.DeviceID = strings.Split(m.Topic(), "/")[2]

	return record, nil
}
