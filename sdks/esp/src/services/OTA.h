#pragma once

#include <stdint.h>
#include <stddef.h>
#include "../codec/CborCodec.h"

class OTA {
public:
    static void begin();
    static void markValid();

private:
    static void _readPayload(const char* topic, const uint8_t* payload, size_t length);
    static void _handleManifest(const uint8_t* payload, size_t length);
    static void _download(const CborCodec::Manifest& man);
    static bool _attempt(const CborCodec::Manifest& man);

    static void _report(const char* event, const char* fwVersion = nullptr);
    static void _reportError(const char* exception, const char* fwVersion, const char* msg);
};