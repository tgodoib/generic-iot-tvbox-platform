#pragma once

#include <stdint.h>

class Heartbeat {
public:
    static void begin(uint32_t intervalMs = 30000);
    static void loop();
    static void monitor(bool enabled, int64_t timeoutMs);

private:
    static uint32_t      _intervalMs;
    static unsigned long _lastMs;
    static bool          _started;
};