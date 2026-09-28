from typing import Any, Optional

from ssaiot.adapters.broker import Broker


class Sensor:
    def __init__(self, broker: Broker, sensor_id: str) -> None:
        self._broker = broker
        self._sensor_id = sensor_id

    def register(self, measurement: str, value: Any, timestamp: Optional[int] = None, retain: bool = False) -> None:
        payload = {"value": value}
        if timestamp:
            payload["timestamp"] = timestamp
        self._broker.publish(
            self._broker.topics().sensor(self._sensor_id, measurement),
            payload,
            retain=retain,
        )
