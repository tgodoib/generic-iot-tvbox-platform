#ifndef TOTEMFIRMWARE_WATCHDOG_H
#define TOTEMFIRMWARE_WATCHDOG_H

#include <Arduino.h>

#define WATCHDOG_PIN LED_BUILTIN

class WatchDog {
private:
    static long last_pulse_start_ms;
    static long pulse_interval_ms;

    static long pulse_duration_ms;

public:
    static void setup_pins();
    static void pulse();
};


#endif //TOTEMFIRMWARE_WATCHDOG_H