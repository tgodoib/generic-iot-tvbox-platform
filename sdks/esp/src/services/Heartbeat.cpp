#include "Heartbeat.h"
#include "../codec/CborCodec.h"
#include "../transport/MQTT.h"
#include "../transport/Topics.h"

#if defined(BOARD_ESP8266)
  #include <ESP8266WiFi.h>
#elif defined(BOARD_ESP32)
  #include <WiFi.h>
#endif

uint32_t      Heartbeat::_intervalMs = 30000;
unsigned long Heartbeat::_lastMs     = 0;
bool          Heartbeat::_started    = false;

void Heartbeat::begin(uint32_t intervalMs) {
    _intervalMs = intervalMs;
    _started    = true;
    _lastMs     = 0;
}

void Heartbeat::loop() {
    if (!_started) return;

    unsigned long now = millis();
    if (now - _lastMs < _intervalMs) return;
    _lastMs = now;

    uint8_t buf[128];
    auto builder = CborCodec::StateBuilder(buf, sizeof(buf))
        .add("free_heap", (int64_t) ESP.getFreeHeap());

#if defined(BOARD_ESP8266)
    builder
        .add("heap_frag",      (int64_t) ESP.getHeapFragmentation())
        .add("max_free_block", (int64_t) ESP.getMaxFreeBlockSize());
#elif defined(BOARD_ESP32)
    builder
        .add("min_free_heap", (int64_t) ESP.getMinFreeHeap())
        .add("chip_temp",     (float)   temperatureRead());
#endif

    size_t n = builder.encode();

    if (n == 0) {
        Serial.println("[Heartbeat] CBOR encode failed");
        return;
    }

    MQTT::ins().publish(Topics::heartbeat().c_str(), buf, n);
}

void Heartbeat::monitor(bool enabled, int64_t timeout_s) {
    uint8_t buf[32];
    size_t n = CborCodec::encodeMonitor(buf, sizeof(buf), enabled, timeout_s);
    if (n == 0) {
        Serial.println("[Heartbeat] Monitor encode failed");
        return;
    }
    MQTT::ins().publish(Topics::monitor().c_str(), buf, n);
}