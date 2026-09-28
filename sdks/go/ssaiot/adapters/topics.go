package adapters

import "fmt"

type Topics struct {
	deviceID    string
	deviceClass string
}

func (b *Broker) Topics() *Topics {
	return &Topics{
		deviceID:    b.cfg.DeviceID,
		deviceClass: b.cfg.DeviceClass,
	}
}

func (t *Topics) Heartbeat() string {
	return fmt.Sprintf("ssaiot/device/%s/state", t.deviceID)
}

func (t *Topics) Monitor() string {
	return fmt.Sprintf("ssaiot/device/%s/state/monitor", t.deviceID)
}

func (t *Topics) Log() string {
	return fmt.Sprintf("ssaiot/device/%s/log", t.deviceID)
}

func (t *Topics) Sensor(sensorID string, measurement string) string {
	return fmt.Sprintf("ssaiot/device/%s/sensor/%s/%s", t.deviceID, sensorID, measurement)
}

func (t *Topics) OtaIdManifest() string {
	return fmt.Sprintf("ssaiot/ota/device/%s/manifest", t.deviceID)
}

func (t *Topics) OtaClassManifest() string {
	return fmt.Sprintf("ssaiot/ota/class/%s/manifest", t.deviceClass)
}
