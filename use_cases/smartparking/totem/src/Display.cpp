#include "Display.h"

#include <LiquidCrystal.h>
#include <SoftwareSerial.h>

#include "services/Logger.h"

#if defined(BOARD_ESP32)
LiquidCrystal lcd(2, 3, 5, 6, 7, 8);
HardwareSerial espSerial(1);
#else
EspSoftwareSerial::UART espSerial;
#endif

bool Display::is_enabled = true;

void Display::setup_pins() {
    // pinMode(UNITS_A_PIN, OUTPUT);
    // pinMode(UNITS_B_PIN, OUTPUT);
    // pinMode(UNITS_C_PIN, OUTPUT);
    // pinMode(UNITS_D_PIN, OUTPUT);
    //
    // pinMode(TENS_A_PIN, OUTPUT);
    // pinMode(TENS_B_PIN, OUTPUT);
    // pinMode(TENS_C_PIN, OUTPUT);
    // pinMode(TENS_D_PIN, OUTPUT);

#if defined(BOARD_ESP32)
    lcd.begin(16, 2);
    espSerial.begin(9600, SERIAL_8N1, 4, 0); // RX=GPIO4, TX=GPIO0
#else
    espSerial.begin(9600, SWSERIAL_8N1, D2, D5); // RX=D2, TX=D5
#endif
}

void Display::write(unsigned int number) {
    if (!is_enabled) {
        Logger::info("Enabling display.");
    }
    is_enabled = true;

    // unsigned int units = number % 10;
    // unsigned int tens = number / 10;
    //
    // bool units_A = units >> 0 & 1;
    // bool units_B = units >> 1 & 1;
    // bool units_C = units >> 2 & 1;
    // bool units_D = units >> 3 & 1;
    //
    // bool tens_A = tens >> 0 & 1;
    // bool tens_B = tens >> 1 & 1;
    // bool tens_C = tens >> 2 & 1;
    // bool tens_D = tens >> 3 & 1;
    //
    // digitalWrite(UNITS_A_PIN, units_A);
    // digitalWrite(UNITS_B_PIN, units_B);
    // digitalWrite(UNITS_C_PIN, units_C);
    // digitalWrite(UNITS_D_PIN, units_D);
    //
    // digitalWrite(TENS_A_PIN, tens_A);
    // digitalWrite(TENS_B_PIN, tens_B);
    // digitalWrite(TENS_C_PIN, tens_C);
    // digitalWrite(TENS_D_PIN, tens_D);

#if defined(BOARD_ESP32)
    lcd.setCursor(0, 0);
    lcd.print("                ");
    lcd.setCursor(0, 0);
    lcd.print("ESP32: ");
    lcd.print(number);
#elif defined(BOARD_ESP8266)
    espSerial.write((byte)number);
#endif
}

void Display::disable() {
    is_enabled = false;
    // digitalWrite(UNITS_A_PIN, 0);
    // digitalWrite(UNITS_B_PIN, 1);
    // digitalWrite(UNITS_C_PIN, 0);
    // digitalWrite(UNITS_D_PIN, 1);
    //
    // digitalWrite(TENS_A_PIN, 0);
    // digitalWrite(TENS_B_PIN, 1);
    // digitalWrite(TENS_C_PIN, 0);
    // digitalWrite(TENS_D_PIN, 1);

#if defined(BOARD_ESP32)
    lcd.setCursor(0, 0);
    lcd.print("                ");
    lcd.setCursor(0, 0);
    lcd.print("ESP32: ");
    lcd.print(-1);
#elif defined(BOARD_ESP8266)
    espSerial.write((byte)-1);
#endif

}

void Display::check_uart() {
#ifdef BOARD_ESP32
    while (espSerial.available()) {
        byte value = espSerial.read();
        Serial.println("Recebeu do ESP8266: " + String(value));

        lcd.setCursor(0, 1);
        lcd.print("                ");
        lcd.setCursor(0, 1);
        lcd.print("ESP8266: ");
        lcd.print(value);
    }
#endif
}