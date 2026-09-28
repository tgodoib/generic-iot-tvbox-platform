#include "../include/WatchDog.h"

long WatchDog::last_pulse_start_ms = 0;

long WatchDog::pulse_interval_ms = 5 * 1 * 1000;
long WatchDog::pulse_duration_ms = 500;

void WatchDog::setup_pins() {
    pinMode(WATCHDOG_PIN, OUTPUT);
}

void WatchDog::pulse() {

    if (millis() - last_pulse_start_ms >= pulse_interval_ms) {
        digitalWrite(WATCHDOG_PIN, HIGH);
        last_pulse_start_ms = millis();
    }

    if (digitalRead(WATCHDOG_PIN) == HIGH && millis() - last_pulse_start_ms >= pulse_duration_ms) {
        digitalWrite(WATCHDOG_PIN, LOW);
    }
}
