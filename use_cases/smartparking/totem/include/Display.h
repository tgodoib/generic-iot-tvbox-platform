#ifndef TOTEM_DISPLAY_H
#define TOTEM_DISPLAY_H

#include "Arduino.h"

#if defined(ESP8266)

#define TENS_A_PIN D3
#define TENS_B_PIN D0
#define TENS_C_PIN D1
#define TENS_D_PIN D2

#define UNITS_A_PIN D8
#define UNITS_B_PIN D5
#define UNITS_C_PIN D6
#define UNITS_D_PIN D7

#elif defined(ESP32)

#define TENS_A_PIN 4    // D3 → GPIO4
#define TENS_B_PIN 16   // D0 → GPIO16
#define TENS_C_PIN 5    // D1 → GPIO5
#define TENS_D_PIN 18   // D2 → GPIO18

#define UNITS_A_PIN 15  // D8 → GPIO15
#define UNITS_B_PIN 14  // D5 → GPIO14
#define UNITS_C_PIN 12  // D6 → GPIO12
#define UNITS_D_PIN 13  // D7 → GPIO13

#endif

class Display {
public:

    static bool is_enabled;

    static void setup_pins();

    static void write(unsigned int number);

    static void disable();

    static void check_uart(); // TEMP
};


#endif
