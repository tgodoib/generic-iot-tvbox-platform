#include "Topics.h"
#include "SSAIoT.h"

std::string Topics::heartbeat() {
    return "ssaiot/device/" + SSAIoT::ins()._deviceId + "/state";
}

std::string Topics::monitor() {
    return "ssaiot/device/" + SSAIoT::ins()._deviceId + "/state/monitor";
}

std::string Topics::log() {
    return "ssaiot/device/" + SSAIoT::ins()._deviceId + "/log";
}

std::string Topics::sensor(const std::string& sensorId, const std::string& measurement) {
    return "ssaiot/device/" + SSAIoT::ins()._deviceId + "/sensor/" + sensorId + "/" + measurement;
}

std::string Topics::otaId() {
    return "ssaiot/ota/device/" + SSAIoT::ins()._deviceId + "/manifest";
}

std::string Topics::otaClass() {
    return "ssaiot/ota/class/" + SSAIoT::ins()._deviceClass + "/manifest";
}