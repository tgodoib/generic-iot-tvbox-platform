package monitor

import (
	"fmt"
	"time"

	"github.com/fxamacker/cbor/v2"
)

type AlertRecord struct {
	Title    string `cbor:"title"`
	Body     string `cbor:"body"`
	Type     string `cbor:"type"`
	DeviceID string `cbor:"device_id"`
}

func encode(title, body, deviceID string) ([]byte, error) {
	payload, err := cbor.Marshal(&AlertRecord{
		Title:    title,
		Body:     body,
		Type:     "health",
		DeviceID: deviceID,
	})

	if err != nil {
		return nil, err
	}

	return payload, nil
}

func GetOffAlert(device string, lastHeartbeat time.Time) ([]byte, error) {
	title := fmt.Sprintf("🚨 Device offline!")
	body := fmt.Sprintf("Last heartbeat:\n%s\n\nTimeout:\n%s", lastHeartbeat.Format("02/01/2006 15h04m05s MST"), time.Since(lastHeartbeat).Round(time.Second))

	return encode(title, body, device)
}
func GetOnAlert(device string, lastHeartbeat time.Time) ([]byte, error) {
	title := fmt.Sprintf("✅ Device online!")
	body := fmt.Sprintf("Was offline for:\n%s", time.Since(lastHeartbeat).Round(time.Second))

	return encode(title, body, device)
}
