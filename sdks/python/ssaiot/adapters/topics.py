class Topics:
    def __init__(self, device_id: str, device_class: str) -> None:
        self._device_id = device_id
        self._device_class = device_class

    def heartbeat(self) -> str:
        return f"ssaiot/device/{self._device_id}/state"

    def monitor(self) -> str:
        return f"ssaiot/device/{self._device_id}/state/monitor"

    def log(self) -> str:
        return f"ssaiot/device/{self._device_id}/log"

    def sensor(self, sensor_id: str, measurement: str) -> str:
        return f"ssaiot/device/{self._device_id}/sensor/{sensor_id}/{measurement}"

    def ota_id_manifest(self) -> str:
        return f"ssaiot/ota/device/{self._device_id}/manifest"

    def ota_class_manifest(self) -> str:
        return f"ssaiot/ota/class/{self._device_class}/manifest"
