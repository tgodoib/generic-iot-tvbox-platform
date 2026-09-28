import hashlib
import os
from dataclasses import dataclass
from enum import Enum
from typing import Callable, Optional

import cbor2

import ssaiot.services.log as log
from ssaiot.adapters.broker import Broker
from ssaiot.adapters.web import Web
from ssaiot.config import Config


class Event(str, Enum):
    MANIFEST_RECEIVED = "manifest_received"
    MANIFEST_CORRECT = "manifest_correct"
    FM_RECEIVED = "fm_received"
    CHECKSUM_OK = "checksum_ok"
    UPDATE_DONE = "update_done"
    UPDATE_COMPLETE = "update_complete"


class Exception_(str, Enum):
    MANIFEST_INCORRECT = "manifest_incorrect"
    FM_INCORRECT = "fm_incorrect"
    BOOT_INCORRECT = "boot_incorrect"
    ROLLBACK = "rollback"


@dataclass
class Manifest:
    manifest_version: str
    firmware_version: str
    firmware_uri: str
    checksum: str
    device_class: str


class Ota:
    def __init__(self, cfg: Config, broker: Broker, web: Web, update_handler: Callable[[str], bool]) -> None:
        if not log.is_configured():
            raise RuntimeError("log.configure(broker) must be called before creating Ota")
        self._cfg = cfg
        self._broker = broker
        self._web = web
        self._update_handler = update_handler

    def listen(self) -> None:
        topics = self._broker.topics()
        self._broker.on_message(topics.ota_id_manifest(), lambda p: self._handle_manifest(p))
        self._broker.on_message(topics.ota_class_manifest(), lambda p: self._handle_manifest(p))

    def _report_event(self, event: Event, firmware_version: Optional[str] = None) -> None:
        tags = {"ota": True, "event": event.value}
        if firmware_version:
            tags["firmware_version"] = firmware_version
        if event is Event.UPDATE_COMPLETE:
            log.info("Update successful!", tags)
        else:
            log.debug("", tags)

    def _report_exception(self, exc: Exception_, firmware_version: Optional[str] = None, msg: str = "") -> None:
        tags = {"ota": True, "exception": exc.value}
        if firmware_version:
            tags["firmware_version"] = firmware_version
        log.error(msg, tags)

    def _handle_manifest(self, payload: bytes) -> None:
        self._report_event(Event.MANIFEST_RECEIVED)

        try:
            manifest = Manifest(**cbor2.loads(payload))
        except Exception:
            self._report_exception(Exception_.MANIFEST_INCORRECT, msg=f"Hex manifest: {payload.hex()}")
            return

        if manifest.device_class != self._cfg.device_class:
            self._report_exception(
                Exception_.MANIFEST_INCORRECT,
                firmware_version=manifest.firmware_version,
                msg=f"Incompatible device class: am {self._cfg.device_class} received {manifest.device_class}",
            )
            return

        self._report_event(Event.MANIFEST_CORRECT, firmware_version=manifest.firmware_version)
        self._download_firmware(manifest)

    def _download_firmware(self, manifest: Manifest) -> None:
        try:
            response = self._web.get(manifest.firmware_uri)
        except Exception as e:
            self._report_exception(Exception_.FM_INCORRECT, manifest.firmware_version,
                                   f"Error downloading firmware: {e}")
            return

        file_path = f"./{manifest.firmware_version}.fm"
        try:
            f = open(file_path, "wb")
        except Exception as e:
            self._report_exception(Exception_.FM_INCORRECT, manifest.firmware_version, f"Error creating firmware file: {e}")
            return

        self._report_event(Event.FM_RECEIVED, manifest.firmware_version)

        hasher = hashlib.sha256()
        try:
            with f:
                while chunk := response.read(8192):
                    f.write(chunk)
                    hasher.update(chunk)
        except Exception as e:
            os.remove(file_path)
            self._report_exception(Exception_.FM_INCORRECT, manifest.firmware_version, f"Error writing firmware file: {e}")
            return

        digest = hasher.hexdigest()
        if digest != manifest.checksum:
            os.remove(file_path)
            self._report_exception(
                Exception_.FM_INCORRECT,
                manifest.firmware_version,
                f"Checksum does not match: calculated {digest} received {manifest.checksum}",
            )
            return

        self._report_event(Event.CHECKSUM_OK, manifest.firmware_version)
        self._update(manifest, file_path)

    def _update(self, manifest: Manifest, file_path: str) -> None:
        if not self._update_handler(file_path):
            self._report_exception(Exception_.BOOT_INCORRECT, manifest.firmware_version, "Update handler failed")
            return
        self._report_event(Event.UPDATE_DONE, manifest.firmware_version)
        self._report_event(Event.UPDATE_COMPLETE, manifest.firmware_version)
