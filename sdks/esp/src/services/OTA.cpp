#include "OTA.h"
#include "Logger.h"
#include "Sha256.h"
#include "Storage.h"
#include "../SSAIoT.h"
#include "../transport/MQTT.h"
#include "../transport/Topics.h"
#include "../transport/SecureClient.h"

#if defined(BOARD_ESP32)
#include <HTTPClient.h>
#include <Update.h>
#include <esp_ota_ops.h>
#elif defined(BOARD_ESP8266)
#include <ESP8266HTTPClient.h>
#include <Updater.h>
#endif

static inline void updateAbort() {
#if defined(BOARD_ESP32)
    Update.abort();
#elif defined(BOARD_ESP8266)
    // UpdaterClass has no abort(). end(false) discards the in-progress update
    // and frees the internal ~4 KB write buffer, preventing heap exhaustion
    // if OTA is retried in the same session.
    Update.end(false);
#endif
}

// ── Event / exception string constants (match Go server) ─────────────────────

static constexpr const char *EV_MANIFEST_RECEIVED = "manifest_received";
static constexpr const char *EV_MANIFEST_CORRECT = "manifest_correct";
static constexpr const char *EV_FM_RECEIVED = "fm_received";
static constexpr const char *EV_CHECKSUM_OK = "checksum_ok";
static constexpr const char *EV_UPDATE_DONE = "update_done";
static constexpr const char *EV_UPDATE_COMPLETE = "update_complete";

static constexpr const char *EX_MANIFEST_INCORRECT = "manifest_incorrect";
static constexpr const char *EX_FM_INCORRECT = "fm_incorrect";
// EXCEPTION_BOOT_INCORRECT     Exception = "boot_incorrect"
// EXCEPTION_ROLLBACK           Exception = "rollback"
// TODO: check for those exceptions and report them

static constexpr size_t kChunkSize   = 4096;
static constexpr int    kMaxAttempts = 3;

// ── Public ────────────────────────────────────────────────────────────────────

void OTA::begin() {
    MQTT::ins().subscribe(
        Topics::otaClass().c_str(),
        [](const char *t, const uint8_t *p, size_t l) { _readPayload(t, p, l); }
    );
    MQTT::ins().subscribe(
        Topics::otaId().c_str(),
        [](const char *t, const uint8_t *p, size_t l) { _readPayload(t, p, l); }
    );

#if defined(BOARD_ESP32)
    {
        const esp_partition_t *p = esp_ota_get_running_partition();
        Serial.printf("[Boot] partition=%s subtype=0x%02x, fwver=%s, otarbt=%d\n",
            p->label, p->subtype,
            Storage::readFWVersion().c_str(),
            Storage::wasRebootForUpdate());
    }
#endif

    if (Storage::wasRebootForUpdate()) {
        Storage::clearRebootForUpdate();
        markValid();
    } else if (Storage::readFWVersion().empty()) {
        Storage::writeFWVersion(SSAIOT_FW_VERSION);
    }
}

void OTA::markValid() {
#if defined(BOARD_ESP32)
    esp_ota_mark_app_valid_cancel_rollback();
#endif
    _report(EV_UPDATE_COMPLETE, Storage::readFWVersion().c_str());
}

// ── Private ───────────────────────────────────────────────────────────────────

void OTA::_readPayload(const char * /*topic*/, const uint8_t *payload, size_t length) {
    _handleManifest(payload, length);
}

void OTA::_handleManifest(const uint8_t *payload, size_t length) {
    Serial.print("[OTA] manifest bytes="); Serial.println(length);
    Serial.print("[OTA] hex: ");
    for (size_t i = 0; i < length && i < 80; i++) {
        if (payload[i] < 0x10) Serial.print('0');
        Serial.print(payload[i], HEX); Serial.print(' ');
    }
    if (length > 80) Serial.println("...");
    else Serial.println();

    // Decode before any MQTT publish: _report() calls MQTT::publish() which
    // overwrites PubSubClient's RX buffer — the same memory payload points into.
    CborCodec::Manifest man;
    bool decoded = CborCodec::decodeManifest(payload, length, man);

    _report(EV_MANIFEST_RECEIVED);  // safe now; payload no longer needed

    if (!decoded) {
        _reportError(EX_MANIFEST_INCORRECT, "", "CBOR decode failed");
        return;
    }

    if (SSAIoT::ins()._deviceClass != man.deviceClass) {
        _reportError(EX_MANIFEST_INCORRECT, man.firmwareVersion, "Device class mismatch");
        return;
    }

    _report(EV_MANIFEST_CORRECT, man.firmwareVersion);

    Logger::debug("Checking firmware version...", [&](auto &b) {
        b.tag("current", Storage::readFWVersion().c_str()).tag("received", man.firmwareVersion);
    });

    if (man.firmwareVersion == Storage::readFWVersion()) {
        Logger::debug("Already up to date!");
        return;
    }

    _download(man);
}

void OTA::_download(const CborCodec::Manifest &man) {
    for (int attempt = 1; attempt <= kMaxAttempts; attempt++) {
        if (attempt > 1) {
            Serial.print("[OTA] Retry "); Serial.print(attempt); Serial.print("/"); Serial.println(kMaxAttempts);
            delay(3000);
        }
        SecureClient::ins().http().stop();
        if (_attempt(man)) return;
    }
    _reportError(EX_FM_INCORRECT, man.firmwareVersion, "Download failed");
}

bool OTA::_attempt(const CborCodec::Manifest &man) {
    HTTPClient http;
    http.begin(SecureClient::ins().http(), man.firmwareUri);

    Serial.println(man.firmwareUri);

    int code = http.GET();
    if (code < 0) {
        // Transport/TLS error — retriable
        Serial.print("[OTA] GET failed: "); Serial.println(code);
        http.end();
        return false;
    }
    if (code != 200) {
        http.end();
        char msg[32];
        snprintf(msg, sizeof(msg), "HTTP %d", code);
        _reportError(EX_FM_INCORRECT, man.firmwareVersion, msg);
        return true;
    }

    int contentLength = http.getSize();
    Serial.print("[OTA] Content-Length: "); Serial.println(contentLength);
    if (contentLength <= 0) {
        http.end();
        _reportError(EX_FM_INCORRECT, man.firmwareVersion, "No Content-Length in response");
        return true;
    }

    if (!Update.begin((size_t) contentLength)) {
        http.end();
        Serial.print("[OTA] Update.begin failed, free heap: "); Serial.println(ESP.getFreeHeap());
        _reportError(EX_FM_INCORRECT, man.firmwareVersion, "Update.begin failed");
        return true;
    }

    WiFiClient *stream = http.getStreamPtr();
    stream->setTimeout(30000);
    Serial.print("[OTA] stream available before loop="); Serial.println(stream->available());

    Sha256 sha256;
    static uint8_t chunk[kChunkSize];
    int remaining = contentLength;
    int totalRead = 0;
    unsigned long dlStart = millis();

    while (remaining > 0) {
        unsigned long chunkStart = millis();
        size_t toRead = (size_t) min((int) sizeof(chunk), remaining);
        size_t n = stream->readBytes(chunk, toRead);
        unsigned long chunkMs = millis() - chunkStart;

        if (n == 0) {
            Serial.print("[OTA] readBytes returned 0 after ");
            Serial.print(chunkMs);
            Serial.print("ms, connected=");
            Serial.print(stream->connected());
            Serial.print(", rssi="); Serial.print(WiFi.RSSI());
            Serial.print(", heap="); Serial.print(ESP.getFreeHeap());
            Serial.print(", remaining="); Serial.println(remaining);
            updateAbort();
            http.end();
            return false;
        }

        if (chunkMs > 500) {
            Serial.print("[OTA] ");
            Serial.print(n); Serial.print("B in ");
            Serial.print(chunkMs); Serial.print("ms, rssi=");
            Serial.print(WiFi.RSSI()); Serial.print(", read=");
            Serial.print(totalRead); Serial.print(", remaining=");
            Serial.println(remaining);
        }

        sha256.update(chunk, n);
        if (Update.write(chunk, n) != n) {
            updateAbort();
            http.end();
            _reportError(EX_FM_INCORRECT, man.firmwareVersion, "Flash write error");
            return true;
        }
        remaining  -= (int) n;
        totalRead  += (int) n;
    }

    unsigned long dlMs = millis() - dlStart;
    Serial.print("[OTA] Download done — read="); Serial.print(totalRead);
    Serial.print(", remaining="); Serial.print(remaining);
    Serial.print(", elapsed="); Serial.print(dlMs); Serial.println("ms");

    http.end();

    _report(EV_FM_RECEIVED, man.firmwareVersion);

    char computed[65];
    sha256.finishHex(computed);

    if (strcmp(computed, man.checksum) != 0) {
        updateAbort();
        _reportError(EX_FM_INCORRECT, man.firmwareVersion, "Checksum mismatch");
        return true;
    }

    _report(EV_CHECKSUM_OK, man.firmwareVersion);

    if (!Update.end(true)) {
        _reportError(EX_FM_INCORRECT, man.firmwareVersion, "Update.end failed");
        return true;
    }

    _report(EV_UPDATE_DONE, man.firmwareVersion);

    delay(500);
    Storage::writeFWVersion(man.firmwareVersion);
    Storage::setRebootForUpdate();
    ESP.restart();
    return true;
}

// ── Reporting ─────────────────────────────────────────────────────────────────

void OTA::_report(const char *event, const char *fwVersion) {
    const char *msg = (strcmp(event, EV_UPDATE_COMPLETE) == 0) ? "Update successful!" : "Update progress...";
    auto tags = [&](auto &b) {
        b.tag("ota", true).tag("event", event);
        if (fwVersion && fwVersion[0]) b.tag("firmware_version", fwVersion);
    };
    if (strcmp(event, EV_UPDATE_COMPLETE) == 0)
        Logger::info(msg, tags);
    else
        Logger::debug(msg, tags);
}

void OTA::_reportError(const char *exception, const char *fwVersion, const char *msg) {
    Logger::error(msg, [&](auto &b) {
        b.tag("ota", true).tag("exception", exception);
        if (fwVersion && fwVersion[0]) b.tag("firmware_version", fwVersion);
    });
}
