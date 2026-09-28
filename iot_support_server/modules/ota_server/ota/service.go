package ota

import (
	"net/url"
	"ota_server/model"
	"strings"
)

type Service struct {
	serverDomain string
	storage      Storage
	broker       Broker
}

func NewService(domain string, storage Storage, broker Broker) *Service {
	return &Service{serverDomain: domain, storage: storage, broker: broker}
}

func (s *Service) ListFirmware() (model.FirmwareList, error) {
	return s.storage.ListFirmware()
}

func (s *Service) CreateFirmware(firmware model.Firmware) (string, error) {
	return s.storage.StoreFirmware(firmware)
}

func (s *Service) DeleteFirmware(id string) error {
	return s.storage.DeleteFirmware(id)
}

func (s *Service) GetFirmwareData(id string) (model.FirmwareMeta, error) {
	return s.storage.GetFirmwareData(id)
}

func (s *Service) CreateManifest(id string) (model.Manifest, error) {
	data, err := s.storage.GetFirmwareData(id)
	if err != nil {
		return model.Manifest{}, err
	}

	uri, err := s.storage.GenerateLink(id)
	if err != nil {
		return model.Manifest{}, err
	}

	parsed, err := url.Parse(uri.String())
	if err != nil {
		return model.Manifest{}, err
	}

	manifest := model.Manifest{
		ManifestVersion: "1.0",
		FirmwareVersion: data.Version,
		FirmwareUri:     "https://" + s.serverDomain + strings.Replace(parsed.RequestURI(), "/firmware", "/firmware/download", 1),
		Checksum:        data.Checksum,
		DeviceClass:     data.DeviceClass,
	}

	return manifest, nil
}

func (s *Service) SendUpdate(id string, recipient model.Recipient) error {
	manifest, err := s.CreateManifest(id)

	if err != nil {
		return err
	}

	if recipient.Type == model.RecipientTypeID {
		err = s.broker.SendManifestToID(recipient.Value, manifest)

	} else {
		err = s.broker.SendManifestToClass(recipient.Value, manifest)
	}

	return err
}
