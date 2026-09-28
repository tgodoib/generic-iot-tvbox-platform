#pragma once

#include <string>
#include <stdint.h>
#include "../codec/CborCodec.h"

class Sensor {
public:
    explicit Sensor(std::string sensorId);

    void record(const std::string& measurement, float       value, int64_t timestamp = 0) const;
    void record(const std::string& measurement, int64_t     value, int64_t timestamp = 0) const;
    void record(const std::string& measurement, bool        value, int64_t timestamp = 0) const;
    void record(const std::string& measurement, const char* value, int64_t timestamp = 0) const;

private:
    std::string _sensorId;
};
