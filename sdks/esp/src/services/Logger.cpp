#include "Logger.h"
#include "../transport/MQTT.h"
#include "../transport/Topics.h"

void Logger::begin(unsigned long baud) {
    Serial.begin(baud);
}

void Logger::_log(int level, const char* msg) {
    Serial.print('['); Serial.print(LogLevel::toString(level)); Serial.print("] ");
    Serial.println(msg ? msg : "");

    uint8_t buf[kLogBufSize];
    size_t n = CborCodec::encodeLog(buf, sizeof(buf), level, msg);
    if (n == 0) { Serial.println("[Logger] encode failed"); return; }
    _publish(buf, n);
}

void Logger::_publish(const uint8_t* buf, size_t n) {
    MQTT::ins().publish(Topics::log().c_str(), buf, n);
}