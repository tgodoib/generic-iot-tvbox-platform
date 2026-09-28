import os
import ssl
import tempfile

from ssaiot.config import Config


def make_tls(cfg: Config) -> ssl.SSLContext:
    # Server certificate is verified against the system CA store (Let's Encrypt)
    ctx = ssl.create_default_context()

    with tempfile.NamedTemporaryFile(mode="w", suffix=".crt", delete=False) as f:
        f.write(cfg.client_crt)
        crt_path = f.name
    with tempfile.NamedTemporaryFile(mode="w", suffix=".key", delete=False) as f:
        f.write(cfg.client_key)
        key_path = f.name
    try:
        ctx.load_cert_chain(crt_path, key_path)
    finally:
        os.unlink(crt_path)
        os.unlink(key_path)

    return ctx
