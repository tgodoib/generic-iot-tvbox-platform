package ota

import (
	"net/url"
	"ota_server/model"
)

type Storage interface {
	ListFirmware() (model.FirmwareList, error)
	StoreFirmware(firmware model.Firmware) (string, error)
	DeleteFirmware(id string) error
	GetFirmwareData(id string) (model.FirmwareMeta, error)
	GenerateLink(id string) (*url.URL, error)
}

type Broker interface {
	SendManifestToClass(class string, manifest model.Manifest) error
	SendManifestToID(id string, manifest model.Manifest) error
}
