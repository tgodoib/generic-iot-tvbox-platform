import logging
import os
import threading
from typing import Optional

import psutil

from ssaiot.adapters.broker import Broker

_logger = logging.getLogger(__name__)


class Health:
    def __init__(self, broker: Broker, beat_interval: float, timeout_s: int) -> None:
        self._broker = broker
        self._beat_interval = beat_interval
        self._timeout_s = timeout_s
        self._stop = threading.Event()
        self._thread: Optional[threading.Thread] = None

    def enable_monitor(self) -> None:
        self._broker.publish(
            self._broker.topics().monitor(),
            {"enabled": True, "timeout": self._timeout_s},
        )

    def disable_monitor(self) -> None:
        self._broker.publish(
            self._broker.topics().monitor(),
            {"enabled": False},
        )

    def start(self) -> None:
        self._stop.clear()
        self._thread = threading.Thread(target=self._run, daemon=True)
        self._thread.start()

    def stop(self) -> None:
        self._stop.set()
        if self._thread:
            self._thread.join()

    def _run(self) -> None:
        process = psutil.Process(os.getpid())
        while not self._stop.wait(self._beat_interval):
            try:
                self._broker.publish(
                    self._broker.topics().heartbeat(),
                    {"mem_used": process.memory_info().rss},
                )
            except Exception as e:
                _logger.error("Could not publish heartbeat: %s", e)
