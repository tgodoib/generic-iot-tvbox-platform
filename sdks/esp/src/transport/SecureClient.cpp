#include "SecureClient.h"
#include "../SSAIoT.h"

void SecureClient::setup() {
    const char* caCert     = SSAIoT::ins()._caCert;
    const char* clientCert = SSAIoT::ins()._clientCert;
    const char* clientKey  = SSAIoT::ins()._clientKey;

    // ── NTP sync (required for cert validation) ───────────────────────────────
#if defined(BOARD_ESP8266)
    configTime("BRT3", "pool.ntp.org");
#elif defined(BOARD_ESP32)
    configTime(-3 * 3600, 0, "pool.ntp.org");
#endif
    while (time(nullptr) < 100000) {
        delay(50);
    }

#if SSAIOT_DEBUG
    Serial.print("[SecureClient] Time synced: ");
    Serial.println((long)time(nullptr));
    Serial.print("[SecureClient] Heap before certs: ");
    Serial.println(ESP.getFreeHeap());
#endif

    // ── Load certificates ─────────────────────────────────────────────────────
#if defined(BOARD_ESP8266)
    _ca         = new X509List(caCert);
    _clientCert = new X509List(clientCert);
    _clientKey  = new PrivateKey(clientKey);

    _mqttClient.setTrustAnchors(_ca);
    _httpClient.setTrustAnchors(_ca);

    _mqttClient.setClientECCert(_clientCert, _clientKey, BR_KEYTYPE_EC | BR_KEYTYPE_SIGN, BR_EC_secp256r1);
    _httpClient.setClientECCert(_clientCert, _clientKey, BR_KEYTYPE_EC | BR_KEYTYPE_SIGN, BR_EC_secp256r1);

    _mqttClient.setBufferSizes(512, 512);
    // nginx ssl_buffer_size must be set to 4k on the server — BearSSL drops
    // any TLS record larger than this buffer, stalling readBytes() mid-download.
    _httpClient.setBufferSizes(4096, 512);

    // Reuse TLS session on MQTT reconnects → avoids full handshake (~4 KB RAM peak).
    _mqttClient.setSession(&_mqttSession);
    

#elif defined(BOARD_ESP32)
    _mqttClient.setCACert(caCert);
    _httpClient.setCACert(caCert);

    _mqttClient.setCertificate(clientCert);
    _mqttClient.setPrivateKey(clientKey);

    _httpClient.setCertificate(clientCert);
    _httpClient.setPrivateKey(clientKey);
#endif

#if SSAIOT_DEBUG
    Serial.print("[SecureClient] Heap after certs: ");
    Serial.println(ESP.getFreeHeap());
    Serial.println("[SecureClient] Ready.");
#endif

    _ready = true;
}
