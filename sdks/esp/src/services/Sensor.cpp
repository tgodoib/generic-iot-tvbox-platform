#include "Sensor.h"
#include "../transport/MQTT.h"
#include "../transport/Topics.h"

Sensor::Sensor(std::string sensorId) : _sensorId(std::move(sensorId)) {}

static void publish(const std::string& sensorId, const std::string& measurement, const uint8_t* buf, size_t n) {
    if (n == 0) {
        Serial.println("[Sensor] CBOR encode failed");
        return;
    }
    MQTT::ins().publish(Topics::sensor(sensorId, measurement).c_str(), buf, n);
}

void Sensor::record(const std::string& measurement, float value, int64_t timestamp) const {
    uint8_t buf[32];
    publish(_sensorId, measurement, buf, CborCodec::encodeSensor(buf, sizeof(buf), value, timestamp));
}

void Sensor::record(const std::string& measurement, int64_t value, int64_t timestamp) const {
    uint8_t buf[32];
    publish(_sensorId, measurement, buf, CborCodec::encodeSensor(buf, sizeof(buf), value, timestamp));
}

void Sensor::record(const std::string& measurement, bool value, int64_t timestamp) const {
    uint8_t buf[16];
    publish(_sensorId, measurement, buf, CborCodec::encodeSensor(buf, sizeof(buf), value, timestamp));
}

void Sensor::record(const std::string& measurement, const char* value, int64_t timestamp) const {
    uint8_t buf[128];
    publish(_sensorId, measurement, buf, CborCodec::encodeSensor(buf, sizeof(buf), value, timestamp));
}
