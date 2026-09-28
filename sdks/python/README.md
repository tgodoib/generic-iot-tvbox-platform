# SSAIOT Python SDK

Python SDK for the SSAIOT IoT platform. Devices communicate with the server over MQTT using mutual TLS (mTLS).

## Requirements

- Python 3.9+
- A device certificate and private key issued by the SSAIOT CA

## Setup

### Using uv (recommended)

```bash
# Create virtual environment
uv venv

# Install the package and its dependencies
uv pip install -e .
```

### Using pip

```bash
# Create virtual environment
python -m venv .venv
source .venv/bin/activate 

# Install the package and its dependencies
pip install -e .
```

## Running

Edit `main.py` with your device credentials, then:

```bash
python main.py
```

The process connects to the broker and blocks, listening for OTA manifests. Uncomment the `Health` and `Sensor` blocks in `main.py` to enable those services.

## Project structure

```
ssaiot/
  config.py          # Config dataclass (device_id, device_class, server_addr, certs)
  adapters/
    _tls.py          # mTLS context builder, shared by broker and web
    broker.py        # MQTT client (paho-mqtt, CBOR encoding, per-topic dispatch)
    topics.py        # MQTT topic string generation
    web.py           # HTTPS client for firmware downloads
  services/
    log.py           # Module-level structured logger — call log.configure(broker) first
    sensor.py        # Publishes sensor measurements
    health.py        # Periodic heartbeat + monitor toggle
    ota.py           # OTA firmware update handler
```