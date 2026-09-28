#pragma once

#include <Arduino.h>

#if defined(BOARD_ESP32)
#include <WiFi.h>
#elif defined(BOARD_ESP8266)
#include <ESP8266WiFi.h>
#endif

class Net {
public:
    static void begin(const char* ssid, const char* password);
    static void loop();

private:
    static const char* _ssid;
    static const char* _password;

    static void _connect();
};