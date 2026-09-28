#include "SSAIoT.h"

#include "services/Storage.h"

void SSAIoT::config(const std::string& addr,
                    const std::string& deviceId,
                    const std::string& deviceClass,
                    const char* caCert,
                    const char* clientCert,
                    const char* clientKey) {
    _addr        = addr;
    _deviceId    = deviceId;
    _deviceClass = deviceClass;
    _caCert      = caCert;
    _clientCert  = clientCert;
    _clientKey   = clientKey;

#ifdef SSAIOT_FW_VERSION
    if (!Storage::wasRebootForUpdate()) {
        Storage::writeFWVersion(std::string(SSAIOT_FW_VERSION));
    }
#endif
}
