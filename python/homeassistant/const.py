"""Constants for the demeter integration."""

from __future__ import annotations

from datetime import timedelta
from enum import StrEnum

DOMAIN = "demeter"

# How often the demeter API is polled for the latest readings.
UPDATE_INTERVAL = timedelta(seconds=60)


class DeviceKind(StrEnum):
    """Kinds of demeter devices that expose metrics."""

    ENCLOSURE = "enclosure"
    SYSTEM = "system"
    PLANT_SITE = "plant_site"


class Metric(StrEnum):
    """Metrics exposed by the demeter API."""

    AIR_TEMPERATURE = "air_temperature"
    AIR_HUMIDITY = "air_humidity"
    PH = "ph"
    EC = "ec"
    WATER_TEMPERATURE = "water_temperature"
    LUX = "lux"


def reading_key(kind: DeviceKind, device_id: int, metric: Metric) -> str:
    """Build the key under which a reading is stored in coordinator data."""
    return f"{kind.value}:{device_id}:{metric.value}"
