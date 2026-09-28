from typing import IO
from urllib.request import urlopen

from ssaiot.adapters._tls import make_tls
from ssaiot.config import Config


class Web:
    def __init__(self, cfg: Config) -> None:
        self._tls = make_tls(cfg)

    def get(self, url: str) -> IO[bytes]:
        return urlopen(url, context=self._tls)

