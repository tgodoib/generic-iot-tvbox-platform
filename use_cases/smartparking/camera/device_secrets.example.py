# Template for device_secrets.py (not tracked by git).
# Copy to device_secrets.py and fill in the server domain and the device's mTLS credentials:
#   cp device_secrets.example.py device_secrets.py

SERVER_ADDR = "<server domain>"

CLIENT_KEY = (
    "-----BEGIN EC PRIVATE KEY-----\n"
    "<device private key>\n"
    "-----END EC PRIVATE KEY-----"
)

CLIENT_CRT = (
    "-----BEGIN CERTIFICATE-----\n"
    "<device certificate>\n"
    "-----END CERTIFICATE-----"
)
