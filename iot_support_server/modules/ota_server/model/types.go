package model

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type Firmware struct {
	FirmwareMeta
	FirmwareFile
}

type FirmwareMeta struct {
	DeviceClass  string    `json:"device_class"`
	Version      string    `json:"version"`
	Checksum     string    `json:"checksum"`
	LastModified time.Time `json:"last_modified"`
}
type FirmwareFile struct {
	Bin  io.Reader
	Size int64
}

type FirmwareList map[string][]FirmwareListEntry

type FirmwareListEntry struct {
	Version      string    `json:"version"`
	LastModified time.Time `json:"last_modified"`
}

type Manifest struct {
	ManifestVersion string `cbor:"manifest_version"`
	FirmwareVersion string `cbor:"firmware_version"`
	FirmwareUri     string `cbor:"firmware_uri"`
	Checksum        string `cbor:"checksum"`
	DeviceClass     string `cbor:"device_class"`
}

type RecipientType string

const (
	RecipientTypeClass RecipientType = "class"
	RecipientTypeID    RecipientType = "id"
)

func (r *RecipientType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch RecipientType(s) {
	case RecipientTypeClass, RecipientTypeID:
		*r = RecipientType(s)
		return nil
	default:
		return fmt.Errorf("invalid subscriptions type: %s", s)
	}
}

type Recipient struct {
	Type  RecipientType `json:"type"`
	Value string        `json:"value"`
}
