import logging
from enum import IntEnum
from typing import Any, Dict, Optional

from ssaiot.adapters.broker import Broker

_logger = logging.getLogger(__name__)
_broker: Optional[Broker] = None


class Level(IntEnum):
    INFO = 0
    DEBUG = 1
    ERROR = 2


def configure(broker: Broker) -> None:
    global _broker
    _broker = broker


def is_configured() -> bool:
    return _broker is not None


def info(msg: str, tags: Optional[Dict[str, Any]] = None) -> None:
    _logger.info("%s %s", tags or {}, msg)
    _send(Level.INFO, msg, tags or {})


def debug(msg: str, tags: Optional[Dict[str, Any]] = None) -> None:
    _logger.debug("%s %s", tags or {}, msg)
    _send(Level.DEBUG, msg, tags or {})


def error(msg: str, tags: Optional[Dict[str, Any]] = None) -> None:
    _logger.error("%s %s", tags or {}, msg)
    _send(Level.ERROR, msg, tags or {})


def _send(level: Level, msg: str, tags: Dict[str, Any]) -> None:
    if _broker is None:
        return
    try:
        _broker.publish(
            _broker.topics().log(),
            {"level": int(level), "tags": tags, "msg": msg},
        )
    except Exception as e:
        _logger.error("Could not send log to server: %s", e)
