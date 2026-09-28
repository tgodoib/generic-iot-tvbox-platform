from dataclasses import dataclass


@dataclass(frozen=True)
class Config:
    device_id: str
    device_class: str
    server_addr: str
    client_key: str
    client_crt: str
