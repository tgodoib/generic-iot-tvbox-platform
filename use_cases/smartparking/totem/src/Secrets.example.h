// Template for Secrets.h (not tracked by git).
// Copy to Secrets.h and fill in the server domain, WiFi and mTLS credentials:
//   cp src/Secrets.example.h src/Secrets.h
#pragma once

#define SERVER_ADDR   "<server domain>"

#define WIFI_SSID     "<wifi ssid>"
#define WIFI_PASSWORD "<wifi password>"

  static const char* CA_CERT = R"(
-----BEGIN CERTIFICATE-----
<CA certificate>
-----END CERTIFICATE-----
)";

  static const char* CLIENT_CERT = R"(
-----BEGIN CERTIFICATE-----
<device certificate>
-----END CERTIFICATE-----
  )";

  static const char* CLIENT_KEY = R"(
-----BEGIN EC PRIVATE KEY-----
<device private key>
-----END EC PRIVATE KEY-----
  )";
