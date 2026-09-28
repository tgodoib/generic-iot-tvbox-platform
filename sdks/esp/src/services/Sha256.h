#pragma once

#include <stdint.h>
#include <stddef.h>

#if defined(BOARD_ESP32)
  #include <mbedtls/sha256.h>
#elif defined(BOARD_ESP8266)
  #include <bearssl/bearssl_hash.h>
#endif

// Streaming SHA-256 hasher. Feed chunks via update(), then call finishHex()
// once to get the lowercase hex digest for comparison with Manifest.checksum.
class Sha256 {
public:
    static constexpr size_t DIGEST_SIZE = 32;

    Sha256() {
#if defined(BOARD_ESP32)
        mbedtls_sha256_init(&_ctx);
        mbedtls_sha256_starts_ret(&_ctx, 0); // 0 = SHA-256 (not SHA-224)
#elif defined(BOARD_ESP8266)
        br_sha256_init(&_ctx);
#endif
    }

    ~Sha256() {
#if defined(BOARD_ESP32)
        mbedtls_sha256_free(&_ctx);
#endif
    }

    void update(const uint8_t* data, size_t len) {
#if defined(BOARD_ESP32)
        mbedtls_sha256_update_ret(&_ctx, data, len);
#elif defined(BOARD_ESP8266)
        br_sha256_update(&_ctx, data, len);
#endif
    }

    // Writes 32 raw bytes. Destructive on ESP32 — do not call update() after.
    void finish(uint8_t out[DIGEST_SIZE]) {
#if defined(BOARD_ESP32)
        mbedtls_sha256_finish_ret(&_ctx, out);
#elif defined(BOARD_ESP8266)
        br_sha256_out(&_ctx, out);
#endif
    }

    // Writes 64 lowercase hex chars + null terminator into out[65].
    void finishHex(char out[65]) {
        uint8_t digest[DIGEST_SIZE];
        finish(digest);
        static const char hex[] = "0123456789abcdef";
        for (size_t i = 0; i < DIGEST_SIZE; i++) {
            out[2*i]   = hex[digest[i] >> 4];
            out[2*i+1] = hex[digest[i] & 0x0f];
        }
        out[64] = '\0';
    }

private:
    Sha256(const Sha256&)            = delete;
    Sha256& operator=(const Sha256&) = delete;

#if defined(BOARD_ESP32)
    mbedtls_sha256_context _ctx;
#elif defined(BOARD_ESP8266)
    br_sha256_context _ctx;
#endif
};
