package record

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

type AlertRecord struct {
	Title    string `cbor:"title"`
	Body     string `cbor:"body"`
	Type     string `cbor:"type"`
	DeviceID string `cbor:"device_id"`
}

func IsOTA(rec *Record) bool {
	_, ok := rec.Tags["ota"]
	return ok
}

func GetAlertRecord(rec *Record) ([]byte, error) {
	title := fmt.Sprintf("Error updating device: %s", rec.DeviceID)
	body := fmt.Sprintf("Firmware: %s \n Exception: %s \n\n Message: \n %s", rec.Tags["firmware_version"], rec.Tags["exception"], rec.Msg)

	payload, err := cbor.Marshal(&AlertRecord{
		Title:    title,
		Body:     body,
		Type:     "ota",
		DeviceID: rec.DeviceID,
	})

	if err != nil {
		return nil, err
	}

	return payload, nil
}
