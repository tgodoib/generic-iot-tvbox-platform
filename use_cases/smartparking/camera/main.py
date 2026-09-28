import json
import logging
import time

import redis

from ssaiot.adapters.broker import Broker
from ssaiot.adapters.web import Web
from ssaiot.config import Config
from ssaiot.services import log
from ssaiot.services.ota import Ota
from ssaiot.services.health import Health
from ssaiot.services.sensor import Sensor

from device_secrets import CLIENT_CRT, CLIENT_KEY, SERVER_ADDR
from inference import SmartParkingCamera

logging.basicConfig(level=logging.DEBUG)


class SSAIoTLogHandler(logging.Handler):
    def emit(self, record: logging.LogRecord) -> None:
        msg = self.format(record)
        if record.levelno >= logging.ERROR:
            log.error(msg)
        elif record.levelno >= logging.INFO:
            log.info(msg)
        else:
            log.debug(msg)


def main() -> None:
    cfg = Config(
        device_id="camera_rpi3",
        device_class="raspberrypi3",
        server_addr=SERVER_ADDR,
        client_key=CLIENT_KEY,
        client_crt=CLIENT_CRT,
    )

    broker = Broker(cfg)
    broker.connect()

    web = Web(cfg)

    log.configure(broker)
    logging.getLogger("inference").addHandler(SSAIoTLogHandler())

    health = Health(broker, beat_interval=1, timeout_s=5)
    # health.enable_monitor()
    health.start()
    health.disable_monitor()

    def handle_update(firmware_path: str) -> bool:
        log.info(f"Applying firmware from {firmware_path}")
        return True

    ota = Ota(cfg, broker, web, update_handler=handle_update)
    ota.listen()

    sensor = Sensor(broker, "camera_01")

    camera = SmartParkingCamera(
        "./assets/yolo11m_float16.tflite",
        "./assets/spots_coordinates.csv",
        "./assets/mask_original_img_768_1024_bw_jul.jpg",
    )

    BACKLOG_QUEUE = "spots_backlog"
    # sensor_data_collector drops anything above 5 msgs/min per sensor; the loop runs
    # at most once a minute, so 3 backlog entries + the live result stay under the limit
    BACKLOG_BATCH = 3
    try:
        redis_client = redis.Redis(host="localhost", port=6379, db=0, decode_responses=True)
        redis_client.ping()
    except redis.RedisError as e:
        log.error(f"Failed to connect to Redis: {e}")
        redis_client = None

    log.info("Service is online, starting inference loop...")

    while True:
        # Check redis backlog queue
        if redis_client is not None and broker.is_connected():
            try:
                for raw in redis_client.lrange(BACKLOG_QUEUE, 0, BACKLOG_BATCH - 1):
                    entry = json.loads(raw)
                    sensor.register("spots", entry["bitmask"], entry["timestamp"], True)
                    redis_client.lpop(BACKLOG_QUEUE)
            except redis.RedisError as e:
                log.error(f"Redis error while flushing backlog: {e}")

        # Run inference
        result = camera.run_inference()
        if result is not None:
            timestamp = int(time.time())
            print(timestamp)
            if broker.is_connected():
                sensor.register("spots", result.bitmask, timestamp, True)
            else:
                log.error("Broker not connected, adding result to backlog.")
                try:
                    if redis_client is not None:
                        redis_client.rpush(BACKLOG_QUEUE, json.dumps({"bitmask": result.bitmask, "timestamp": timestamp}))
                except redis.RedisError as e:
                    log.error(f"Redis error while pushing to backlog: {e}")

        time.sleep(60)


if __name__ == "__main__":
    main()
