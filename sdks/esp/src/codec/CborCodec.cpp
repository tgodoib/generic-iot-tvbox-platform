#include "CborCodec.h"

#include <HardwareSerial.h>
#include <string.h>

namespace CborCodec {
    // ── Manifest decoder ──────────────────────────────────────────────────────────

    bool decodeManifest(const uint8_t *buf, size_t len, Manifest &out) {
        memset(&out, 0, sizeof(out));

        CborParser parser;
        CborValue root, map;

        if (cbor_parser_init(buf, len, 0, &parser, &root) != CborNoError) {
            Serial.println("[CBOR] init failed"); return false;
        }
        if (!cbor_value_is_map(&root)) {
            Serial.println("[CBOR] root not a map"); return false;
        }
        if (cbor_value_enter_container(&root, &map) != CborNoError) {
            Serial.println("[CBOR] enter container failed"); return false;
        }

        while (!cbor_value_at_end(&map)) {
            if (!cbor_value_is_text_string(&map)) {
                Serial.println("[CBOR] key not a string"); return false;
            }

            char key[32];
            size_t keyLen = sizeof(key);
            if (cbor_value_copy_text_string(&map, key, &keyLen, &map) != CborNoError) {
                Serial.println("[CBOR] key copy failed"); return false;
            }
            key[sizeof(key) - 1] = '\0';

            if (!cbor_value_is_text_string(&map)) {
                Serial.print("[CBOR] value for '"); Serial.print(key); Serial.println("' not a string"); return false;
            }

            char *dst = nullptr;
            size_t capacity = 0;

            if (strcmp(key, "manifest_version") == 0) {
                dst = out.manifestVersion;
                capacity = sizeof(out.manifestVersion);
            } else if (strcmp(key, "firmware_version") == 0) {
                dst = out.firmwareVersion;
                capacity = sizeof(out.firmwareVersion);
            } else if (strcmp(key, "firmware_uri") == 0) {
                dst = out.firmwareUri;
                capacity = sizeof(out.firmwareUri);
            } else if (strcmp(key, "checksum") == 0) {
                dst = out.checksum;
                capacity = sizeof(out.checksum);
            } else if (strcmp(key, "device_class") == 0) {
                dst = out.deviceClass;
                capacity = sizeof(out.deviceClass);
            }

            if (dst) {
                size_t avail = capacity;
                CborError err = cbor_value_copy_text_string(&map, dst, &avail, &map);
                if (err != CborNoError) {
                    Serial.print("[CBOR] copy '"); Serial.print(key);
                    Serial.print("' failed err="); Serial.println(err); return false;
                }
                dst[capacity - 1] = '\0';
            } else {
                if (cbor_value_advance(&map) != CborNoError) {
                    Serial.print("[CBOR] advance after '"); Serial.print(key); Serial.println("' failed"); return false;
                }
            }
        }

        if (cbor_value_leave_container(&root, &map) != CborNoError) {
            Serial.println("[CBOR] leave container failed"); return false;
        }
        return true;
    }

    // ── LogRecord ─────────────────────────────────────────────────────────────────

    size_t encodeLog(uint8_t *buf, size_t cap, int level, const char *msg, int64_t timestamp) {
        int fields = 1;
        if (msg && msg[0]) fields++;
        if (timestamp) fields++;

        CborEncoder root, map;
        cbor_encoder_init(&root, buf, cap, 0);
        CborError err = cbor_encoder_create_map(&root, &map, fields);

        if (err == CborNoError) err = cbor_encode_text_stringz(&map, "level");
        if (err == CborNoError) err = cbor_encode_int(&map, level);

        if (msg && msg[0]) {
            if (err == CborNoError) err = cbor_encode_text_stringz(&map, "msg");
            if (err == CborNoError) err = cbor_encode_text_stringz(&map, msg);
        }

        if (timestamp) {
            if (err == CborNoError) err = cbor_encode_text_stringz(&map, "timestamp");
            if (err == CborNoError) err = cbor_encode_int(&map, timestamp);
        }

        if (err == CborNoError) err = cbor_encoder_close_container(&root, &map);
        return err == CborNoError ? cbor_encoder_get_buffer_size(&root, buf) : 0;
    }

    // ── SensorRecord ──────────────────────────────────────────────────────────────

    size_t encodeSensor(uint8_t *buf, size_t cap, float value, int64_t timestamp) {
        int fields = 1 + (timestamp ? 1 : 0);
        CborEncoder root, map;
        cbor_encoder_init(&root, buf, cap, 0);
        CborError err = cbor_encoder_create_map(&root, &map, fields);

        if (err == CborNoError) err = cbor_encode_text_stringz(&map, "value");
        if (err == CborNoError) err = cbor_encode_float(&map, value);

        if (timestamp) {
            if (err == CborNoError) err = cbor_encode_text_stringz(&map, "timestamp");
            if (err == CborNoError) err = cbor_encode_int(&map, timestamp);
        }

        if (err == CborNoError) err = cbor_encoder_close_container(&root, &map);
        return err == CborNoError ? cbor_encoder_get_buffer_size(&root, buf) : 0;
    }

    size_t encodeSensor(uint8_t *buf, size_t cap, int64_t value, int64_t timestamp) {
        int fields = 1 + (timestamp ? 1 : 0);
        CborEncoder root, map;
        cbor_encoder_init(&root, buf, cap, 0);
        CborError err = cbor_encoder_create_map(&root, &map, fields);

        if (err == CborNoError) err = cbor_encode_text_stringz(&map, "value");
        if (err == CborNoError) err = cbor_encode_int(&map, value);

        if (timestamp) {
            if (err == CborNoError) err = cbor_encode_text_stringz(&map, "timestamp");
            if (err == CborNoError) err = cbor_encode_int(&map, timestamp);
        }

        if (err == CborNoError) err = cbor_encoder_close_container(&root, &map);
        return err == CborNoError ? cbor_encoder_get_buffer_size(&root, buf) : 0;
    }

    size_t encodeSensor(uint8_t *buf, size_t cap, bool value, int64_t timestamp) {
        int fields = 1 + (timestamp ? 1 : 0);
        CborEncoder root, map;
        cbor_encoder_init(&root, buf, cap, 0);
        CborError err = cbor_encoder_create_map(&root, &map, fields);

        if (err == CborNoError) err = cbor_encode_text_stringz(&map, "value");
        if (err == CborNoError) err = cbor_encode_boolean(&map, value);

        if (timestamp) {
            if (err == CborNoError) err = cbor_encode_text_stringz(&map, "timestamp");
            if (err == CborNoError) err = cbor_encode_int(&map, timestamp);
        }

        if (err == CborNoError) err = cbor_encoder_close_container(&root, &map);
        return err == CborNoError ? cbor_encoder_get_buffer_size(&root, buf) : 0;
    }

    size_t encodeSensor(uint8_t *buf, size_t cap, const char *value, int64_t timestamp) {
        int fields = 1 + (timestamp ? 1 : 0);
        CborEncoder root, map;
        cbor_encoder_init(&root, buf, cap, 0);
        CborError err = cbor_encoder_create_map(&root, &map, fields);

        if (err == CborNoError) err = cbor_encode_text_stringz(&map, "value");
        if (err == CborNoError) err = cbor_encode_text_stringz(&map, value);

        if (timestamp) {
            if (err == CborNoError) err = cbor_encode_text_stringz(&map, "timestamp");
            if (err == CborNoError) err = cbor_encode_int(&map, timestamp);
        }

        if (err == CborNoError) err = cbor_encoder_close_container(&root, &map);
        return err == CborNoError ? cbor_encoder_get_buffer_size(&root, buf) : 0;
    }

    static bool decodeSensorCommon(const uint8_t *buf, size_t len, CborValue &valueOut, int64_t &timestamp) {
        timestamp = 0;

        CborParser parser;
        CborValue root, map;
        if (cbor_parser_init(buf, len, 0, &parser, &root) != CborNoError) return false;
        if (!cbor_value_is_map(&root)) return false;
        if (cbor_value_enter_container(&root, &map) != CborNoError) return false;

        bool valueFound = false;

        while (!cbor_value_at_end(&map)) {
            if (!cbor_value_is_text_string(&map)) return false;

            char key[16];
            size_t keyLen = sizeof(key);
            if (cbor_value_copy_text_string(&map, key, &keyLen, &map) != CborNoError) return false;
            key[sizeof(key) - 1] = '\0';

            if (strcmp(key, "value") == 0) {
                valueOut = map;
                valueFound = true;
                if (cbor_value_advance(&map) != CborNoError) return false;
            } else if (strcmp(key, "timestamp") == 0) {
                if (!cbor_value_is_integer(&map)) return false;
                cbor_value_get_int64(&map, &timestamp);
                if (cbor_value_advance(&map) != CborNoError) return false;
            } else {
                if (cbor_value_advance(&map) != CborNoError) return false;
            }
        }

        return valueFound;
    }

    bool decodeSensor(const uint8_t *buf, size_t len, float &value, int64_t &timestamp) {
        CborValue v;
        if (!decodeSensorCommon(buf, len, v, timestamp)) return false;
        if (!cbor_value_is_float(&v)) return false;
        float f;
        cbor_value_get_float(&v, &f);
        value = f;
        return true;
    }

    bool decodeSensor(const uint8_t *buf, size_t len, int64_t &value, int64_t &timestamp) {
        CborValue v;
        if (!decodeSensorCommon(buf, len, v, timestamp)) return false;
        if (!cbor_value_is_integer(&v)) return false;
        cbor_value_get_int64(&v, &value);
        return true;
    }

    bool decodeSensor(const uint8_t *buf, size_t len, bool &value, int64_t &timestamp) {
        CborValue v;
        if (!decodeSensorCommon(buf, len, v, timestamp)) return false;
        if (!cbor_value_is_boolean(&v)) return false;
        cbor_value_get_boolean(&v, &value);
        return true;
    }

    bool decodeSensor(const uint8_t *buf, size_t len, char *value, size_t cap, int64_t &timestamp) {
        CborValue v;
        if (!decodeSensorCommon(buf, len, v, timestamp)) return false;
        if (!cbor_value_is_text_string(&v)) return false;
        if (cbor_value_copy_text_string(&v, value, &cap, nullptr) != CborNoError) return false;
        value[cap - 1] = '\0';
        return true;
    }

    // ── LogBuilder ────────────────────────────────────────────────────────────────

    LogBuilder::LogBuilder(uint8_t *buf, size_t cap)
        : _buf(buf), _root{}, _outer{}, _tags{}, _err(CborNoError), _tagsOpen(false) {
        cbor_encoder_init(&_root, _buf, cap, 0);
        _err = cbor_encoder_create_map(&_root, &_outer, CborIndefiniteLength);
    }

    LogBuilder &LogBuilder::level(int lvl) {
        if (_err == CborNoError) _err = cbor_encode_text_stringz(&_outer, "level");
        if (_err == CborNoError) _err = cbor_encode_int(&_outer, lvl);
        return *this;
    }

    LogBuilder &LogBuilder::msg(const char *m) {
        if (_err == CborNoError) _err = cbor_encode_text_stringz(&_outer, "msg");
        if (_err == CborNoError) _err = cbor_encode_text_stringz(&_outer, m);
        return *this;
    }

    LogBuilder &LogBuilder::ts(int64_t timestamp) {
        if (_err == CborNoError) _err = cbor_encode_text_stringz(&_outer, "timestamp");
        if (_err == CborNoError) _err = cbor_encode_int(&_outer, timestamp);
        return *this;
    }

    void LogBuilder::_ensureTagsOpen() {
        if (_tagsOpen) return;
        if (_err == CborNoError) _err = cbor_encode_text_stringz(&_outer, "tags");
        if (_err == CborNoError) _err = cbor_encoder_create_map(&_outer, &_tags, CborIndefiniteLength);
        _tagsOpen = true;
    }

    LogBuilder &LogBuilder::tag(const char *key, float value) {
        _ensureTagsOpen();
        if (_err == CborNoError) _err = cbor_encode_text_stringz(&_tags, key);
        if (_err == CborNoError) _err = cbor_encode_float(&_tags, value);
        return *this;
    }

    LogBuilder &LogBuilder::tag(const char *key, int64_t value) {
        _ensureTagsOpen();
        if (_err == CborNoError) _err = cbor_encode_text_stringz(&_tags, key);
        if (_err == CborNoError) _err = cbor_encode_int(&_tags, value);
        return *this;
    }

    LogBuilder &LogBuilder::tag(const char *key, bool value) {
        _ensureTagsOpen();
        if (_err == CborNoError) _err = cbor_encode_text_stringz(&_tags, key);
        if (_err == CborNoError) _err = cbor_encode_boolean(&_tags, value);
        return *this;
    }

    LogBuilder &LogBuilder::tag(const char *key, const char *value) {
        _ensureTagsOpen();
        if (_err == CborNoError) _err = cbor_encode_text_stringz(&_tags, key);
        if (_err == CborNoError) _err = cbor_encode_text_stringz(&_tags, value);
        return *this;
    }

    size_t LogBuilder::encode() {
        if (_tagsOpen && _err == CborNoError)
            _err = cbor_encoder_close_container(&_outer, &_tags);
        if (_err == CborNoError)
            _err = cbor_encoder_close_container(&_root, &_outer);
        return _err == CborNoError ? cbor_encoder_get_buffer_size(&_root, _buf) : 0;
    }

    // ── MonitorRecord ─────────────────────────────────────────────────────────────

    size_t encodeMonitor(uint8_t *buf, size_t cap, bool enabled, int64_t timeout) {
        CborEncoder root, map;
        cbor_encoder_init(&root, buf, cap, 0);
        CborError err = cbor_encoder_create_map(&root, &map, 2);

        if (err == CborNoError) err = cbor_encode_text_stringz(&map, "enabled");
        if (err == CborNoError) err = cbor_encode_boolean(&map, enabled);

        if (err == CborNoError) err = cbor_encode_text_stringz(&map, "timeout");
        if (err == CborNoError) err = cbor_encode_int(&map, timeout);

        if (err == CborNoError) err = cbor_encoder_close_container(&root, &map);
        return err == CborNoError ? cbor_encoder_get_buffer_size(&root, buf) : 0;
    }

    // ── StateRecord ───────────────────────────────────────────────────────────────

    StateBuilder::StateBuilder(uint8_t *buf, size_t cap)
        : _buf(buf), _root{}, _map{}, _err(CborNoError) {
        cbor_encoder_init(&_root, _buf, cap, 0);
        _err = cbor_encoder_create_map(&_root, &_map, CborIndefiniteLength);
    }

    StateBuilder &StateBuilder::ts(int64_t timestamp) {
        if (_err == CborNoError) _err = cbor_encode_text_stringz(&_map, "timestamp");
        if (_err == CborNoError) _err = cbor_encode_int(&_map, timestamp);
        return *this;
    }

    StateBuilder &StateBuilder::add(const char *key, float value) {
        if (_err == CborNoError) _err = cbor_encode_text_stringz(&_map, key);
        if (_err == CborNoError) _err = cbor_encode_float(&_map, value);
        return *this;
    }

    StateBuilder &StateBuilder::add(const char *key, int64_t value) {
        if (_err == CborNoError) _err = cbor_encode_text_stringz(&_map, key);
        if (_err == CborNoError) _err = cbor_encode_int(&_map, value);
        return *this;
    }

    StateBuilder &StateBuilder::add(const char *key, bool value) {
        if (_err == CborNoError) _err = cbor_encode_text_stringz(&_map, key);
        if (_err == CborNoError) _err = cbor_encode_boolean(&_map, value);
        return *this;
    }

    StateBuilder &StateBuilder::add(const char *key, const char *value) {
        if (_err == CborNoError) _err = cbor_encode_text_stringz(&_map, key);
        if (_err == CborNoError) _err = cbor_encode_text_stringz(&_map, value);
        return *this;
    }

    size_t StateBuilder::encode() {
        if (_err == CborNoError) _err = cbor_encoder_close_container(&_root, &_map);
        return _err == CborNoError ? cbor_encoder_get_buffer_size(&_root, _buf) : 0;
    }
} // namespace CborCodec
