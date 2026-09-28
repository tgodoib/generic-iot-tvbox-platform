#pragma once
#include <string>

#if defined(BOARD_ESP32)
#include <Preferences.h>
#elif defined(BOARD_ESP8266)
#include <EEPROM.h>
#endif

class Storage {
public:
    static std::string readFWVersion();
    static void        writeFWVersion(const std::string& version);

    static bool wasRebootForUpdate();
    static void setRebootForUpdate();
    static void clearRebootForUpdate();

private:
#if defined(BOARD_ESP8266)
    static constexpr int     kSize         = 34;
    static constexpr uint8_t kMagic        = 0xA5;
    static constexpr int     kFwVerOffset  = 1;
    static constexpr int     kFwVerLen     = 32;
    static constexpr int     kRebootOffset = 33;

    static bool _ready;
    static void _init();
#endif
};