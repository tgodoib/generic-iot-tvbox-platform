package services

import (
	"ssaiot_sim/ssaiot/adapters"
)

type Sensor struct {
	broker *adapters.Broker
	id     string
}

type SensorRecord struct {
	Value     interface{} `cbor:"value"`
	Timestamp int64       `cbor:"timestamp,omitempty"`
}

func NewSensor(broker *adapters.Broker, id string) *Sensor {
	return &Sensor{
		broker: broker,
		id:     id,
	}
}

// Register publishes a measurement. timestamp is a Unix time in seconds; 0 lets the server use its reception time.
func (s *Sensor) Register(measurement string, value any, timestamp int64, retain bool) error {
	return s.broker.Publish(s.broker.Topics().Sensor(s.id, measurement), SensorRecord{value, timestamp}, retain)
}
