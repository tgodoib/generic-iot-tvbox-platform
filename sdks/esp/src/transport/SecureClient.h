#pragma once

// ─── Board abstraction ────────────────────────────────────────────────────────
#if defined(BOARD_ESP8266)
  #include <ESP8266WiFi.h>
  #include <WiFiClientSecureBearSSL.h>
  using WiFiClientSecure = BearSSL::WiFiClientSecure;
  using X509List          = BearSSL::X509List;
  using PrivateKey        = BearSSL::PrivateKey;
  using Session           = BearSSL::Session;
#elif defined(BOARD_ESP32)
  #include <WiFi.h>
  #include <WiFiClientSecure.h>
  // ESP32 uses mbedTLS — no BearSSL types
#else
  #error "Define BOARD_ESP8266 or BOARD_ESP32 in build_flags"
#endif

// ─── SecureClient ─────────────────────────────────────────────────────────────
// Singleton that owns the two mTLS WiFiClientSecure instances shared across
// the library (one for MQTT, one for HTTP).  Call SecureClient::instance().setup()
// once in setup(), after WiFi is connected.
//
// Memory note (ESP8266): certs are heap-allocated once and never freed.
// The two WiFiClientSecure objects are stored inside the singleton (static
// storage), so no extra allocation happens after setup().

class SecureClient {
public:
    // Returns the single instance (created on first call).
    static SecureClient& ins() {
        static SecureClient inst;
        return inst;
    }

    // Delete copy / move to enforce singleton semantics.
    SecureClient(const SecureClient&)            = delete;
    SecureClient& operator=(const SecureClient&) = delete;

    // Syncs NTP, loads certs from Secrets.h, configures both TLS clients.
    // Must be called once after WiFi is up.
    void setup();

    // Accessors used by MqttTransport and HttpTransport.
    WiFiClientSecure& mqtt() { return _mqttClient; }
    WiFiClientSecure& http() { return _httpClient; }

    // True after setup() completes successfully.
    bool isReady() const { return _ready; }

private:
    SecureClient() : _ready(false) {}

    WiFiClientSecure _mqttClient;
    WiFiClientSecure _httpClient;

    bool _ready;

#if defined(BOARD_ESP8266)
    // Heap-allocated once in setup() — shared between both clients to avoid
    // duplicating ~2 KB of cert data in RAM.
    X509List*   _ca         = nullptr;
    X509List*   _clientCert = nullptr;
    PrivateKey* _clientKey  = nullptr;
    Session     _mqttSession;           // TLS session reuse → skip full handshake
#endif
};