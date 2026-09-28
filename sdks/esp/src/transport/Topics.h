#pragma once
#include <string>

class Topics {
public:
    static std::string heartbeat();
    static std::string monitor();
    static std::string log();
    static std::string sensor(const std::string& sensorId, const std::string& measurement);
    static std::string otaId();
    static std::string otaClass();
};