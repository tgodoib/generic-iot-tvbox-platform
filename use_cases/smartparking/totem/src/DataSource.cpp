//
// Created by Tiago Bannwart on 16/06/26.
//

#include "DataSource.h"

#include "Display.h"
#include "codec/CborCodec.h"
#include "services/Logger.h"
#include "transport/MQTT.h"

unsigned int DataSource::last_ok_read = 0;
unsigned int DataSource::max_down_time_s = 60 * 5;

void DataSource::setup() {
    MQTT::ins().subscribe(
        "ssaiot/device/camera_rpi3/sensor/camera_01/spots",
        [](const char *t, const uint8_t *p, size_t l) {
            int64_t spots_bitmask;
            int64_t timestamp;

            if (!CborCodec::decodeSensor(p, l, spots_bitmask, timestamp)) {
                Logger::error("Error decoding sensor data from camera_rpi3 camera_01 spots");
                return;
            }

            Serial.println("Valor recebido via MQTT: " + String(spots_bitmask));
            last_ok_read = millis();

            uint64_t spots = spots_bitmask;
            uint64_t mask = 0b111111110001111;
            uint64_t total = 15 - 3;

            uint64_t available = total - __builtin_popcountll(spots & mask);

            Display::write(available);
        });
}

bool DataSource::is_updated() {
    return millis() - last_ok_read <= max_down_time_s * 1000;
}