#include "MQTT.h"
#include <cstring>

#include "SSAIoT.h"

MQTT::MQTT() : _client(SecureClient::ins().mqtt()) {}

// ─── Public ──────────────────────────────────────────────────────────────────

void MQTT::begin() {


    _client.setServer(SSAIoT::ins()._addr.c_str(), _port);
    _client.setBufferSize(1024);
    _client.setKeepAlive(120);
    _client.setCallback(_globalCallback);

    _connect();
}

void MQTT::loop() {
    if (!_client.connected()) {
        unsigned long now = millis();
        if (now - _lastAttemptMs >= _backoffMs) {
            _lastAttemptMs = now;
            _connect();
        }
        return;
    }
    _client.loop();
}

bool MQTT::publish(const char *topic, const uint8_t *payload, size_t length) {
    if (!_client.connected()) {
        Serial.println("[MQTT] publish failed: not connected");
        return false;
    }
    return _client.publish(topic, payload, length);
}

bool MQTT::subscribe(const char *topic, MqttMessageCallback callback) {
    // Look for an existing entry for this filter.
    SubEntry *entry = _findEntry(topic);

    if (entry == nullptr) {
        // New filter — need a free slot.
        if (_subCount >= kMaxSubs) {
            Serial.println("[MQTT] subscribe failed: subscription table full");
            return false;
        }
        entry = &_subs[_subCount++];
        entry->filter        = topic;
        entry->callbackCount = 0;
    }

    // Append the callback if there's room.
    if (entry->callbackCount >= kMaxCallbacksPerSub) {
        Serial.println("[MQTT] subscribe failed: too many callbacks for topic");
        return false;
    }
    entry->callbacks[entry->callbackCount++] = callback;

    if (!_client.connected()) return false;

    bool ok = _client.subscribe(topic);
    return ok;
}

bool MQTT::isConnected() {
    return _client.connected();
}

void MQTT::_connect() {
    if (!SecureClient::ins().isReady()) {
        Serial.println("[MQTT] SecureClient not ready — skipping connect");
        return;
    }

    Serial.print("[MQTT] Connecting to ");
    Serial.println(SSAIoT::ins()._addr.c_str());

    if (_client.connect(SSAIoT::ins()._deviceId.c_str())) {
        Serial.println("[MQTT] Connected.");
        _backoffMs = 1000;
        _resubscribeAll();
    } else {
        Serial.print("[MQTT] Failed, rc=");
        Serial.println(_client.state());
        Serial.print("[MQTT] Retry in ");
        Serial.print(_backoffMs / 1000);
        Serial.println(" s");
        _backoffMs = min(_backoffMs * 2, kMaxBackoffMs);
    }
}

void MQTT::_resubscribeAll() {
    for (uint8_t i = 0; i < _subCount; i++) {
        _client.subscribe(_subs[i].filter.c_str());
        Serial.print("[MQTT] Resubscribed: ");
        Serial.println(_subs[i].filter.c_str());
    }
}

// Static dispatch — called by PubSubClient on every incoming message.
void MQTT::_globalCallback(const char *topic, uint8_t *payload, unsigned int length) {
    // Serial.print("[MQTT] Message on: '");
    // Serial.print(topic);
    // Serial.println("'");

    MQTT &self = ins();

    // Serial.print("[MQTT] Checking ");
    // Serial.print(self._subCount);
    // Serial.println(" filter(s):");
    // for (uint8_t i = 0; i < self._subCount; i++) {
    //     Serial.print("  [");
    //     Serial.print(i);
    //     Serial.print("] '");
    //     Serial.print(self._subs[i].filter.c_str());
    //     Serial.println("'");
    // }

    bool dispatched = false;
    for (uint8_t i = 0; i < self._subCount; i++) {
        if (_topicMatches(self._subs[i].filter.c_str(), topic)) {
            for (uint8_t j = 0; j < self._subs[i].callbackCount; j++) {
                self._subs[i].callbacks[j](topic, payload, length);
            }
            dispatched = true;
        }
    }

    if (!dispatched) {
        Serial.print("[MQTT] No handler for topic: ");
        Serial.println(topic);
    }
}

// Matches a topic filter (with + and # wildcards) against a concrete topic.
//
//   filter:  "ssaiot/+/alert"   topic: "ssaiot/health/alert"  → true
//   filter:  "ssaiot/#"         topic: "ssaiot/ota/device/x"  → true
//   filter:  "ssaiot/+/alert"   topic: "ssaiot/health/data"   → false
//
bool MQTT::_topicMatches(const char *filter, const char *topic) {
    const char *f = filter;
    const char *t = topic;

    while (*f && *t) {
        if (*f == '#') {
            // '#' matches everything that remains (including slashes).
            return true;
        }
        if (*f == '+') {
            // '+' matches exactly one level — advance topic past the next '/'.
            while (*t && *t != '/') ++t;
            ++f; // step past '+'
        } else {
            if (*f != *t) return false;
            ++f;
            ++t;
        }
    }

    // Both exhausted: exact match (or filter ends with '#').
    if (*f == '#') return true;
    return *f == '\0' && *t == '\0';
}

MQTT::SubEntry *MQTT::_findEntry(const char *filter) {
    for (uint8_t i = 0; i < _subCount; i++) {
        if (_subs[i].filter == filter) {
            return &_subs[i];
        }
    }
    return nullptr;
}