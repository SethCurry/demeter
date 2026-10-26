"""Sensor platform for the demeter integration.

Every demeter device that exposes metrics becomes a Home Assistant device
with one sensor entity per metric:

* enclosures: air temperature and air humidity
* systems: pH, EC and water temperature
* plant sites: illuminance (lux)

The coordinator listener adds entities for devices created in demeter
after setup.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING, Any

from homeassistant.components.sensor import (
    SensorDeviceClass,
    SensorEntity,
    SensorStateClass,
)
from homeassistant.const import LIGHT_LUX, PERCENTAGE, UnitOfTemperature
from homeassistant.core import HomeAssistant
from homeassistant.helpers.device_registry import DeviceInfo
from homeassistant.helpers.entity_platform import AddEntitiesCallback
from homeassistant.helpers.update_coordinator import CoordinatorEntity

from . import DemeterConfigEntry
from .const import DOMAIN, DeviceKind, Metric, reading_key
from .coordinator import DemeterCoordinator

if TYPE_CHECKING:
    from .coordinator import Device


@dataclass(frozen=True)
class MetricSpec:
    """Describes the sensor entity exposed for one demeter metric."""

    metric: Metric
    name: str
    device_class: SensorDeviceClass | None = None
    unit: str | None = None
    icon: str | None = None


# NOTE: demeter stores raw EC values without a unit annotation; µS/cm is
# the usual unit for hobby EC sensors. Adjust here if demeter settles on a
# different convention.
DEVICE_METRICS: dict[DeviceKind, tuple[MetricSpec, ...]] = {
    DeviceKind.ENCLOSURE: (
        MetricSpec(
            Metric.AIR_TEMPERATURE,
            "Air Temperature",
            SensorDeviceClass.TEMPERATURE,
            UnitOfTemperature.CELSIUS,
        ),
        MetricSpec(
            Metric.AIR_HUMIDITY,
            "Air Humidity",
            SensorDeviceClass.HUMIDITY,
            PERCENTAGE,
        ),
    ),
    DeviceKind.SYSTEM: (
        MetricSpec(Metric.PH, "pH", icon="mdi:flask"),
        MetricSpec(Metric.EC, "EC", unit="µS/cm", icon="mdi:omega"),
        MetricSpec(
            Metric.WATER_TEMPERATURE,
            "Water Temperature",
            SensorDeviceClass.TEMPERATURE,
            UnitOfTemperature.CELSIUS,
        ),
    ),
    DeviceKind.PLANT_SITE: (
        MetricSpec(
            Metric.LUX,
            "Illuminance",
            SensorDeviceClass.ILLUMINANCE,
            LIGHT_LUX,
        ),
    ),
}


def _device_info(
    coordinator: DemeterCoordinator, kind: DeviceKind, device: Device
) -> DeviceInfo:
    """Build Home Assistant device info for a demeter device."""
    if kind is DeviceKind.PLANT_SITE:
        flow = coordinator.data.flows.get(device.flow_id)
        flow_name = flow.name if flow else "Unknown flow"
        name = f"{flow_name} Site ({device.x}, {device.y}, {device.z})"
        model = "Plant Site"
    elif kind is DeviceKind.ENCLOSURE:
        name = device.name
        model = "Enclosure"
    else:
        name = device.name
        model = "System"
    return DeviceInfo(
        identifiers={(DOMAIN, f"{kind.value}_{device.id}")},
        name=name,
        manufacturer="Demeter",
        model=model,
        configuration_url=coordinator.client.host,
    )


class DemeterSensor(CoordinatorEntity[DemeterCoordinator], SensorEntity):
    """A sensor fed by the latest demeter reading of a single metric."""

    _attr_has_entity_name = True
    _attr_state_class = SensorStateClass.MEASUREMENT

    def __init__(
        self,
        coordinator: DemeterCoordinator,
        kind: DeviceKind,
        device_id: int,
        spec: MetricSpec,
        device_info: DeviceInfo,
    ) -> None:
        """Initialize the sensor."""
        super().__init__(coordinator)
        self._reading_key = reading_key(kind, device_id, spec.metric)
        self._attr_unique_id = f"{kind.value}_{device_id}_{spec.metric.value}"
        self._attr_name = spec.name
        self._attr_device_class = spec.device_class
        self._attr_native_unit_of_measurement = spec.unit
        self._attr_icon = spec.icon
        self._attr_device_info = device_info

    @property
    def native_value(self) -> float | int | None:
        """Return the latest reading value, if demeter has one."""
        reading = self.coordinator.data.readings.get(self._reading_key)
        if reading is None:
            return None
        return reading.value

    @property
    def extra_state_attributes(self) -> dict[str, Any] | None:
        """Expose when the reading was taken."""
        reading = self.coordinator.data.readings.get(self._reading_key)
        if reading is None or reading.timestamp is None:
            return None
        return {"reading_timestamp": reading.timestamp.isoformat()}


async def async_setup_entry(
    hass: HomeAssistant,
    entry: DemeterConfigEntry,
    async_add_entities: AddEntitiesCallback,
) -> None:
    """Set up demeter sensors from a config entry."""
    coordinator = entry.runtime_data.coordinator

    seen: set[str] = set()

    def _add_new_entities() -> None:
        entities: list[DemeterSensor] = []
        data = coordinator.data
        for kind, specs in DEVICE_METRICS.items():
            for device in data.devices(kind):
                for spec in specs:
                    key = reading_key(kind, device.id, spec.metric)
                    if key in seen:
                        continue
                    seen.add(key)
                    entities.append(
                        DemeterSensor(
                            coordinator,
                            kind,
                            device.id,
                            spec,
                            _device_info(coordinator, kind, device),
                        )
                    )
        if entities:
            async_add_entities(entities)

    # Add entities discovered in the first refresh, and any devices demeter
    # gains later.
    entry.async_on_unload(coordinator.async_add_listener(_add_new_entities))
    _add_new_entities()
