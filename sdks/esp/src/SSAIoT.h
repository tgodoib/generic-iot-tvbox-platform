#pragma once

#include <string>

class SSAIoT {
public:
    static SSAIoT& ins() {
        static SSAIoT inst;
        return inst;
    }

    SSAIoT(const SSAIoT&)            = delete;
    SSAIoT& operator=(const SSAIoT&) = delete;

    void config(const std::string& addr,
                const std::string& deviceId,
                const std::string& deviceClass,
                const char* caCert,
                const char* clientCert,
                const char* clientKey);

    std::string _addr;
    std::string _deviceId;
    std::string _deviceClass;

    // PEM strings — point directly to caller-owned literals, not copied.
    const char* _caCert     = nullptr;
    const char* _clientCert = nullptr;
    const char* _clientKey  = nullptr;

private:
    SSAIoT() = default;
};
