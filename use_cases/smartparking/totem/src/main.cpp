#include <Arduino.h>

#include "DataSource.h"
#include "Secrets.h"
#include "Net.h"
#include "Display.h"
#include "WatchDog.h"

#include "SSAIoT.h"
#include "transport/MQTT.h"
#include "services/Logger.h"
#include "services/OTA.h"
#include "services/Heartbeat.h"
#include "services/Storage.h"


void setup() {
    pinMode(LED_BUILTIN, OUTPUT);
    digitalWrite(LED_BUILTIN, HIGH);
    Serial.begin(115200);

    Logger::begin();

    String id = "";
    String dclass = "";

#if defined(BOARD_ESP32)
    id = "totem_esp32";
    dclass = "esp32";
#else
    id = "totem_esp8266";
    dclass = "esp8266";
#endif

    SSAIoT::ins().config(
        SERVER_ADDR,
        id.c_str(),
        dclass.c_str(),
        CA_CERT,
        CLIENT_CERT,
        CLIENT_KEY
    );

    Display::setup_pins();

    Net::begin(WIFI_SSID, WIFI_PASSWORD);
    SecureClient::ins().setup();
    MQTT::ins().begin();

    Heartbeat::begin(30 * 1000);
    Heartbeat::monitor(false, 60);

    OTA::begin();

    DataSource::setup();

    Logger::info("Totem ready, listening to spots data...");
    Logger::debug(("Current firmware version: " + Storage::readFWVersion()).c_str());
}

void loop() {
    Net::loop();
    MQTT::ins().loop();
    Heartbeat::loop();

    if (!DataSource::is_updated() && !Display::is_enabled) {
        Logger::info("New data timeout, disabling display.");
        Display::disable();
    }

    WatchDog::pulse();

    Display::check_uart(); // TEMP
}
