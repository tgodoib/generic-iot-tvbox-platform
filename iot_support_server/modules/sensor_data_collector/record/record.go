package record

import (
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/fxamacker/cbor/v2"
)

type SensorDataRecord struct {
	DeviceID    string
	SensorID    string
	Measurement string
	Timestamp   int64       `cbor:"timestamp,omitempty"`
	Value       interface{} `cbor:"value"`
}

type AlertRecord struct {
	Title    string `cbor:"title"`
	Body     string `cbor:"body"`
	Type     string `cbor:"type"`
	DeviceID string `cbor:"device_id"`
}

func NewSensorDataRecord(m MQTT.Message) (SensorDataRecord, error) {
	var record SensorDataRecord

	err := cbor.Unmarshal(m.Payload(), &record)
	if err != nil {
		return SensorDataRecord{}, err
	}
	record.DeviceID = strings.Split(m.Topic(), "/")[2]
	record.SensorID = strings.Split(m.Topic(), "/")[4]
	record.Measurement = strings.Split(m.Topic(), "/")[5]

	if record.Timestamp == 0 {
		record.Timestamp = time.Now().Unix()
	}

	return record, nil
}

func EncodeAlertRecord(rec *AlertRecord) ([]byte, error) {
	payload, err := cbor.Marshal(rec)
	if err != nil {
		return nil, err
	}
	return payload, nil
}
