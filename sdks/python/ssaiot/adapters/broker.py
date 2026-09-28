from typing import Callable, Dict, List

import cbor2
import paho.mqtt.client as mqtt

from ssaiot.adapters._tls import make_tls
from ssaiot.adapters.topics import Topics
from ssaiot.config import Config

MessageHandler = Callable[[bytes], None]


class Broker:
    def __init__(self, cfg: Config) -> None:
        self._cfg = cfg
        self._handlers: Dict[str, List[MessageHandler]] = {}
        self._client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION2, client_id=cfg.device_id)
        self._client.tls_set_context(make_tls(cfg))
        self._client.on_connect = self._subscribe_all
        self._client.on_message = self._dispatch

    def connect(self) -> None:
        self._client.connect(self._cfg.server_addr, 8883)
        self._client.loop_start()

    def is_connected(self) -> bool:
        return self._client.is_connected()

    def close(self) -> None:
        self._client.loop_stop()
        self._client.disconnect()

    def publish(self, topic: str, payload: object, retain: bool = False) -> None:
        self._client.publish(topic, cbor2.dumps(payload), retain=retain)

    def on_message(self, topic: str, handler: MessageHandler) -> None:
        is_new = topic not in self._handlers
        self._handlers.setdefault(topic, []).append(handler)
        if is_new and self._client.is_connected():
            self._client.subscribe(topic)

    def topics(self) -> Topics:
        return Topics(self._cfg.device_id, self._cfg.device_class)

    def __enter__(self) -> "Broker":
        self.connect()
        return self

    def __exit__(self, *_) -> None:
        self.close()

    # Runs on every (re)connect: clean sessions drop subscriptions on the broker side
    def _subscribe_all(self, client: mqtt.Client, userdata: object, flags: object, reason_code: object, properties: object) -> None:
        if self._handlers:
            client.subscribe([(t, 0) for t in self._handlers])

    def _dispatch(self, client: mqtt.Client, userdata: object, msg: mqtt.MQTTMessage) -> None:
        for handler in self._handlers.get(msg.topic, []):
            handler(msg.payload)
