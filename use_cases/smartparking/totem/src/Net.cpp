#include "Net.h"

const char* Net::_ssid     = nullptr;
const char* Net::_password = nullptr;

void Net::begin(const char* ssid, const char* password) {
    Serial.begin(115200);
    _ssid     = ssid;
    _password = password;
    _connect();
}

void Net::loop() {
    if (WiFi.status() != WL_CONNECTED) {
        Serial.println("\n[Net] Connection lost, reconnecting...");
        _connect();
    }
}

void Net::_connect() {
    Serial.print("[Net] Connecting to ");
    Serial.print(_ssid);
    Serial.print("...");

    WiFi.begin(_ssid, _password);
    while (WiFi.status() != WL_CONNECTED) {
        delay(500);
        Serial.print('.');
    }

    Serial.println();
    Serial.print("[Net] Connected. IP: ");
    Serial.println(WiFi.localIP());
}