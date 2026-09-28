package record

import (
	"fmt"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/fxamacker/cbor/v2"
)

type AlertType string

const (
	AlertAll        AlertType = "*"
	AlertHealth     AlertType = "health"
	AlertOTAError   AlertType = "ota"
	AlertSensorData AlertType = "sensor_data"
)

type AlertRecord struct {
	Title    string    `cbor:"title"`
	Body     string    `cbor:"body"`
	Type     AlertType `cbor:"type"`
	DeviceID string    `cbor:"device_id"`
}

func NewAlertRecord(m MQTT.Message) (AlertRecord, error) {
	var record AlertRecord

	err := cbor.Unmarshal(m.Payload(), &record)
	if err != nil {
		return AlertRecord{}, err
	}

	return record, nil
}

func (r *AlertType) UnmarshalCBOR(data []byte) error {
	var s string
	if err := cbor.Unmarshal(data, &s); err != nil {
		return err
	}
	switch AlertType(s) {
	case AlertAll, AlertHealth, AlertOTAError, AlertSensorData:
		*r = AlertType(s)
		return nil
	default:
		return fmt.Errorf("invalid alert type: %s", s)
	}
}
