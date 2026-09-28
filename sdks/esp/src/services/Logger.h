#pragma once

#include <Arduino.h>
#include <stdint.h>
#include "../codec/CborCodec.h"

static constexpr size_t kLogBufSize = 320;

struct SerialTagWriter {
    bool _first = true;

    SerialTagWriter& tag(const char* k, float v)       { _sep(k); Serial.print(v);               return *this; }
    SerialTagWriter& tag(const char* k, int64_t v)     { _sep(k); Serial.print((long)v);          return *this; }
    SerialTagWriter& tag(const char* k, bool v)        { _sep(k); Serial.print(v?"true":"false"); return *this; }
    SerialTagWriter& tag(const char* k, const char* v) { _sep(k); Serial.print(v);               return *this; }
    void close() { if (!_first) Serial.print('}'); }

private:
    void _sep(const char* k) {
        if (_first) { Serial.print("  {"); _first = false; } else Serial.print(", ");
        Serial.print(k); Serial.print('=');
    }
};

class Logger {
public:
    static void begin(unsigned long baud = 115200);

    static void info(const char* msg)  { _log(LogLevel::INFO,  msg); }
    static void debug(const char* msg) { _log(LogLevel::DEBUG, msg); }
    static void error(const char* msg) { _log(LogLevel::ERROR, msg); }

    template<typename F>
    static void info(const char* msg, F fn)  { _logTagged(LogLevel::INFO,  msg, fn); }
    template<typename F>
    static void debug(const char* msg, F fn) { _logTagged(LogLevel::DEBUG, msg, fn); }
    template<typename F>
    static void error(const char* msg, F fn) { _logTagged(LogLevel::ERROR, msg, fn); }

private:
    static void _log(int level, const char* msg);
    static void _publish(const uint8_t* buf, size_t n);

    template<typename F>
    static void _logTagged(int level, const char* msg, F fn) {
        Serial.print('['); Serial.print(LogLevel::toString(level)); Serial.print("] ");
        Serial.print(msg ? msg : "");
        SerialTagWriter sw;
        fn(sw);
        sw.close();
        Serial.println();

        uint8_t buf[kLogBufSize];
        CborCodec::LogBuilder b(buf, sizeof(buf));
        b.level(level);
        if (msg && msg[0]) b.msg(msg);
        fn(b);
        size_t n = b.encode();
        if (n == 0) { Serial.println("[Logger] encode failed"); return; }
        _publish(buf, n);
    }
};