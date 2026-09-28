# Generic IoT TV Box Platform (SSAIoT)

SSAIoT (Servidor de Suporte a Aplicações IoT) is a support server for IoT applications: devices talk to it over MQTT with mTLS and CBOR payloads to publish sensor data, logs and heartbeats, and to receive OTA firmware updates.

## Repository layout

```
iot_support_server/   Server modules (Go) + release scripts, nginx/mosquitto config and systemd units
sdks/
  esp/                ESP32/ESP8266 client SDK (PlatformIO library)
  python/             Python client SDK
  go/                 Go client SDK + device simulator
grafana/              Grafana app plugin (firmware upload / browse / deploy)
use_cases/
  smartparking/
    totem/            ESP firmware for the parking totem display (links sdks/esp)
    camera/           Raspberry Pi parking camera (links sdks/python)
```

## Configuration

Secrets and deployment-specific values are never committed. Each component ships a template to copy and fill in:

| Component | Template | Local file |
|---|---|---|
| Server | `iot_support_server/releases/config/ssaiot.env.example` | `ssaiot.env` |
| Totem | `use_cases/smartparking/totem/src/Secrets.example.h` | `Secrets.h` |
| Camera | `use_cases/smartparking/camera/device_secrets.example.py` | `device_secrets.py` |
| Go simulator | `sdks/go/secrets.example.go` | `secrets.go` |
