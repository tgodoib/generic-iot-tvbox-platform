#pragma once

#include <stdint.h>
#include <stddef.h>
#include <cbor.h>

namespace LogLevel {
    constexpr int INFO  = 0;
    constexpr int DEBUG = 1;
    constexpr int ERROR = 2;

    inline const char* toString(int level) {
        switch (level) {
            case INFO:  return "INFO";
            case DEBUG: return "DEBUG";
            case ERROR: return "ERROR";
            default:    return "?";
        }
    }
}

namespace CborCodec {

    // ── Manifest (received from server) ──────────────────────────────────────────
    // Decoded from the CBOR payload on the OTA manifest topic.
    struct Manifest {
        char manifestVersion[16];
        char firmwareVersion[32];
        char firmwareUri[512];
        char checksum[65];   // SHA256 hex (64 chars) + null
        char deviceClass[32];
    };

    // Populate out from a CBOR map. Unknown keys are skipped; missing fields
    // are left as empty strings. Returns false if the payload is not a valid map.
    bool decodeManifest(const uint8_t* buf, size_t len, Manifest& out);

    // LogRecord: { level, msg?, timestamp? }
    // Use LogBuilder when you also need tags.
    size_t encodeLog(uint8_t* buf, size_t cap, int level, const char* msg, int64_t timestamp = 0);

    // LogRecord with tags: { level, msg?, timestamp?, tags: { key→val, ... } }
    // Usage:
    //   uint8_t buf[128];
    //   size_t n = CborCodec::LogBuilder(buf, sizeof(buf))
    //                 .level(LogLevel::INFO)
    //                 .msg("boot complete")
    //                 .ts(now)
    //                 .tag("board", "esp32c3")
    //                 .encode();
    class LogBuilder {
    public:
        LogBuilder(uint8_t* buf, size_t cap);

        LogBuilder& level(int lvl);
        LogBuilder& msg(const char* m);
        LogBuilder& ts(int64_t timestamp);
        LogBuilder& tag(const char* key, float       value);
        LogBuilder& tag(const char* key, int64_t     value);
        LogBuilder& tag(const char* key, bool        value);
        LogBuilder& tag(const char* key, const char* value);

        size_t encode();

    private:
        uint8_t*    _buf;
        CborEncoder _root;
        CborEncoder _outer;
        CborEncoder _tags;
        CborError   _err;
        bool        _tagsOpen;

        void _ensureTagsOpen();
    };

    // SensorRecord: { value, timestamp? }
    size_t encodeSensor(uint8_t* buf, size_t cap, float       value, int64_t timestamp = 0);
    size_t encodeSensor(uint8_t* buf, size_t cap, int64_t     value, int64_t timestamp = 0);
    size_t encodeSensor(uint8_t* buf, size_t cap, bool        value, int64_t timestamp = 0);
    size_t encodeSensor(uint8_t* buf, size_t cap, const char* value, int64_t timestamp = 0);

    bool decodeSensor(const uint8_t* buf, size_t len, float&       value, int64_t& timestamp);
    bool decodeSensor(const uint8_t* buf, size_t len, int64_t&     value, int64_t& timestamp);
    bool decodeSensor(const uint8_t* buf, size_t len, bool&        value, int64_t& timestamp);
    bool decodeSensor(const uint8_t* buf, size_t len, char*        value, size_t cap, int64_t& timestamp);

    // MonitorRecord: { enabled, timeout }
    size_t encodeMonitor(uint8_t* buf, size_t cap, bool enabled, int64_t timeout);

    // StateRecord: { timestamp?, key→val, ... }
    // States (cbor:"-") are flattened into the top-level map alongside timestamp.
    // Uses an indefinite-length CBOR map — no need to count fields upfront.
    //
    // Usage:
    //   uint8_t buf[64];
    //   size_t n = CborCodec::StateBuilder(buf, sizeof(buf))
    //                 .ts(now)
    //                 .add("relay", true)
    //                 .encode();
    class StateBuilder {
    public:
        StateBuilder(uint8_t* buf, size_t cap);

        StateBuilder& ts(int64_t timestamp);
        StateBuilder& add(const char* key, float       value);
        StateBuilder& add(const char* key, int64_t     value);
        StateBuilder& add(const char* key, bool        value);
        StateBuilder& add(const char* key, const char* value);

        size_t encode();

    private:
        uint8_t*    _buf;
        CborEncoder _root;
        CborEncoder _map;
        CborError   _err;
    };

} // namespace CborCodec