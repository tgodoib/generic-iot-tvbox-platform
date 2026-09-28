#include "Storage.h"

// ── ESP32 — Preferences (NVS) ─────────────────────────────────────────────────
#if defined(BOARD_ESP32)

static constexpr const char* kNs = "ssaiot";

std::string Storage::readFWVersion() {
    Preferences p;
    p.begin(kNs, true);
    std::string v = p.getString("fwver", "").c_str();
    p.end();
    return v;
}

void Storage::writeFWVersion(const std::string& version) {
    Preferences p;
    p.begin(kNs, false);
    p.putString("fwver", version.c_str());
    p.end();
}

bool Storage::wasRebootForUpdate() {
    Preferences p;
    p.begin(kNs, true);
    bool v = p.getBool("otarbt", false);
    p.end();
    return v;
}

void Storage::setRebootForUpdate() {
    Preferences p;
    p.begin(kNs, false);
    p.putBool("otarbt", true);
    p.end();
}

void Storage::clearRebootForUpdate() {
    Preferences p;
    p.begin(kNs, false);
    p.putBool("otarbt", false);
    p.end();
}

// ── ESP8266 — EEPROM ──────────────────────────────────────────────────────────
#elif defined(BOARD_ESP8266)

bool Storage::_ready = false;

void Storage::_init() {
    if (_ready) return;
    EEPROM.begin(kSize);
    if (EEPROM.read(0) != kMagic) {
        EEPROM.write(0, kMagic);
        for (int i = 1; i < kSize; i++) EEPROM.write(i, 0);
        EEPROM.commit();
    }
    _ready = true;
}

std::string Storage::readFWVersion() {
    _init();
    char buf[kFwVerLen];
    for (int i = 0; i < kFwVerLen; i++)
        buf[i] = (char) EEPROM.read(kFwVerOffset + i);
    buf[kFwVerLen - 1] = '\0';
    return std::string(buf);
}

void Storage::writeFWVersion(const std::string& version) {
    _init();
    size_t n = version.size() < (size_t)(kFwVerLen - 1) ? version.size() : (kFwVerLen - 1);
    for (size_t i = 0; i < n; i++)
        EEPROM.write(kFwVerOffset + i, (uint8_t) version[i]);
    EEPROM.write(kFwVerOffset + n, 0);
    EEPROM.commit();
}

bool Storage::wasRebootForUpdate() {
    _init();
    return EEPROM.read(kRebootOffset) == 1;
}

void Storage::setRebootForUpdate() {
    _init();
    EEPROM.write(kRebootOffset, 1);
    EEPROM.commit();
}

void Storage::clearRebootForUpdate() {
    _init();
    EEPROM.write(kRebootOffset, 0);
    EEPROM.commit();
}

#endif