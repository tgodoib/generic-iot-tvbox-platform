package services

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	url2 "net/url"
	"os"
	"ssaiot_sim/ssaiot"
	"ssaiot_sim/ssaiot/adapters"

	"github.com/fxamacker/cbor/v2"
)

type Event string
type Exception string

const (
	EVENT_MANIFEST_RECEIVED Event = "manifest_received"
	EVENT_MANIFEST_CORRECT  Event = "manifest_correct"
	EVENT_FM_RECEIVED       Event = "fm_received"
	EVENT_CHECKSUM_OK       Event = "checksum_ok"
	EVENT_UPDATE_DONE       Event = "update_done"
	EVENT_UPDATE_COMPLETE   Event = "update_complete"
)

const (
	EXCEPTION_MANIFEST_INCORRECT Exception = "manifest_incorrect"
	EXCEPTION_FM_INCORRECT       Exception = "fm_incorrect"
	EXCEPTION_BOOT_INCORRECT     Exception = "boot_incorrect"
	EXCEPTION_ROLLBACK           Exception = "rollback"
)

type Manifest struct {
	ManifestVersion string `cbor:"manifest_version"`
	FirmwareVersion string `cbor:"firmware_version"`
	FirmwareUri     string `cbor:"firmware_uri"`
	Checksum        string `cbor:"checksum"`
	DeviceClass     string `cbor:"device_class"`
}

// UpdateHandler applies the downloaded firmware at filePath and reports whether it succeeded.
type UpdateHandler func(filePath string) bool

type Ota struct {
	cfg           *ssaiot.Config
	broker        *adapters.Broker
	web           *adapters.Web
	log           *Log
	updateHandler UpdateHandler
}

func NewOta(cfg *ssaiot.Config, broker *adapters.Broker, web *adapters.Web, log *Log, updateHandler UpdateHandler) *Ota {
	return &Ota{
		cfg:           cfg,
		broker:        broker,
		web:           web,
		log:           log,
		updateHandler: updateHandler,
	}
}

func (o *Ota) Listen() {
	o.broker.OnMessage(o.broker.Topics().OtaIdManifest(), o.handleManifest)
	o.broker.OnMessage(o.broker.Topics().OtaClassManifest(), o.handleManifest)
}

func (o *Ota) reportMessage(ev Event, fmVersion string) {
	tags := map[string]interface{}{
		"ota":   true,
		"event": string(ev),
	}
	if fmVersion != "" {
		tags["firmware_version"] = fmVersion
	}

	if ev == EVENT_UPDATE_COMPLETE {
		o.log.Info("Update successful!", tags)
	} else {
		o.log.Debug("", tags)
	}
}

func (o *Ota) reportException(ex Exception, fmVersion string, err string) {
	tags := map[string]interface{}{
		"ota":       true,
		"exception": string(ex),
	}
	if fmVersion != "" {
		tags["firmware_version"] = fmVersion
	}

	o.log.Error(err, tags)
}

func (o *Ota) handleManifest(payload []byte) {
	o.reportMessage(EVENT_MANIFEST_RECEIVED, "")

	man := new(Manifest)
	err := cbor.Unmarshal(payload, man)
	if err != nil {
		o.reportException(EXCEPTION_MANIFEST_INCORRECT, "", "Hex Manifest: "+hex.EncodeToString(payload))
		return
	} else if man.DeviceClass != o.cfg.DeviceClass {
		o.reportException(EXCEPTION_MANIFEST_INCORRECT, man.FirmwareVersion, "Incompatible device class: am "+o.cfg.DeviceClass+" received "+man.DeviceClass)
		return
	}

	o.reportMessage(EVENT_MANIFEST_CORRECT, man.FirmwareVersion)
	o.downloadFirmware(man)
}

func (o *Ota) downloadFirmware(man *Manifest) {
	url, err := url2.Parse(man.FirmwareUri)
	if err != nil {
		o.reportException(EXCEPTION_FM_INCORRECT, man.FirmwareVersion, "Error parsing uri: "+err.Error())
		return
	}

	payload, err := o.web.Get(url.String())
	if err != nil {
		o.reportException(EXCEPTION_FM_INCORRECT, man.FirmwareVersion, "Error downloading firmware: "+err.Error())
		return
	}
	defer payload.Close()

	filePath := "./" + man.FirmwareVersion + ".fm"
	file, err := os.Create(filePath)
	if err != nil {
		o.reportException(EXCEPTION_FM_INCORRECT, man.FirmwareVersion, "Error creating firmware file: "+err.Error())
		return
	}

	o.reportMessage(EVENT_FM_RECEIVED, man.FirmwareVersion)

	h := sha256.New()
	w := io.MultiWriter(file, h)

	_, err = io.Copy(w, payload)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(filePath)
		o.reportException(EXCEPTION_FM_INCORRECT, man.FirmwareVersion, "Error writing firmware file: "+err.Error())
		return
	}

	hash := hex.EncodeToString(h.Sum(nil))
	if hash != man.Checksum {
		os.Remove(filePath)
		o.reportException(EXCEPTION_FM_INCORRECT, man.FirmwareVersion, "Checksum does not match: calculated "+hash+" received "+man.Checksum)
		return
	}

	o.reportMessage(EVENT_CHECKSUM_OK, man.FirmwareVersion)

	o.update(man, filePath)
}

func (o *Ota) update(man *Manifest, filePath string) {
	if !o.updateHandler(filePath) {
		o.reportException(EXCEPTION_BOOT_INCORRECT, man.FirmwareVersion, "Update handler failed")
		return
	}
	o.reportMessage(EVENT_UPDATE_DONE, man.FirmwareVersion)
	o.reportMessage(EVENT_UPDATE_COMPLETE, man.FirmwareVersion)
}
