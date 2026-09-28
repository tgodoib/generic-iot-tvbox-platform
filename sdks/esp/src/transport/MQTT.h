#pragma once

#include <functional>
#include <PubSubClient.h>
#include "SecureClient.h"
#include "SSAIoT.h"

// Callback type for incoming MQTT messages.
// Accepts free functions, lambdas, and member functions (via std::bind or capturing lambda).
//
// Free function:     void onMsg(const char*, const uint8_t*, size_t) { ... }
//                    mqtt.subscribe("topic", onMsg);
//
// Member function:   mqtt.subscribe("topic",
//                        [this](const char* t, const uint8_t* p, size_t l){
//                            this->onMsg(t, p, l);
//                        });
//                    // or:
//                    mqtt.subscribe("topic",
//                        std::bind(&OTA::onMsg, this,
//                            std::placeholders::_1,
//                            std::placeholders::_2,
//                            std::placeholders::_3));
using MqttMessageCallback = std::function<void(const char *topic, const uint8_t *payload, size_t length)>;

// ─── MQTT ─────────────────────────────────────────────────────────────────────
// Thin wrapper around PubSubClient that:
//   • uses the SecureClient singleton's WiFiClientSecure
//   • handles reconnection with exponential back-off
//   • routes incoming messages to per-topic callbacks (wildcard-aware)
//   • supports multiple callbacks per topic (up to kMaxCallbacksPerSub)
//   • must have loop() called from the Arduino loop() — or via SSAIoTClient
//
// Usage:
//   MQTT::ins().begin();
//   MQTT::ins().subscribe("ssaiot/+/alert",  onAlert);
//   MQTT::ins().subscribe("ssaiot/+/alert",  onAlertLog);  // second callback, same topic
//   MQTT::ins().subscribe("ssaiot/ota/#",    onOta);
//   // in loop():
//   MQTT::ins().loop();

class MQTT {
public:
    static MQTT &ins() {
        static MQTT inst;
        return inst;
    }

    MQTT(const MQTT &) = delete;
    MQTT &operator=(const MQTT &) = delete;

    void begin();
    void loop();

    bool publish(const char *topic, const uint8_t *payload, size_t length);

    // Subscribe to a topic and register a callback to handle incoming messages.
    // The same topic filter can be passed multiple times with different callbacks —
    // all registered callbacks will be invoked on a matching message.
    // Returns false if the subscription table is full or the broker call fails.
    bool subscribe(const char *topic, MqttMessageCallback callback);

    bool isConnected();

private:
    MQTT();

    void _connect();
    void _resubscribeAll();

    // Global PubSubClient callback — dispatches to per-topic callbacks.
    static void _globalCallback(const char *topic, uint8_t *payload, unsigned int length);

    // Returns true if `topic` (concrete, from broker) matches `filter` (may contain + or #).
    static bool _topicMatches(const char *filter, const char *topic);

    PubSubClient _client;
    uint16_t _port = 8883;

    // Back-off state
    unsigned long _lastAttemptMs = 0;
    uint32_t _backoffMs          = 1000;
    static constexpr uint32_t kMaxBackoffMs = 60000;

    // ── Subscription table ───────────────────────────────────────────────────
    // Each entry holds one topic filter and up to kMaxCallbacksPerSub callbacks.
    // Multiple subscribe() calls for the same filter simply add to its callback list.

    static constexpr uint8_t kMaxSubs             = 8;
    static constexpr uint8_t kMaxCallbacksPerSub  = 4;

    struct SubEntry {
        std::string          filter;
        MqttMessageCallback  callbacks[kMaxCallbacksPerSub];
        uint8_t              callbackCount;
    };

    SubEntry _subs[kMaxSubs] = {};
    uint8_t  _subCount       = 0;

    // Finds an existing entry for `filter`, or returns nullptr if not found.
    SubEntry *_findEntry(const char *filter);
};