"""Data update coordinator for the demeter integration.

On every update the coordinator discovers the full demeter topology
(enclosures, systems, flows and plant sites) and pulls the latest reading
of every metric those devices expose.
"""

from __future__ import annotations

import asyncio
import logging
from collections.abc import Awaitable, Callable
from dataclasses import dataclass, field
from typing import TYPE_CHECKING

from homeassistant.core import HomeAssistant
from homeassistant.helpers.update_coordinator import (
    DataUpdateCoordinator,
    UpdateFailed,
)

from .api import (
    DemeterClient,
    DemeterError,
    Enclosure,
    Flow,
    PlantSite,
    Reading,
    System,
)
from .const import DOMAIN, UPDATE_INTERVAL, DeviceKind, Metric, reading_key

if TYPE_CHECKING:
    from . import DemeterConfigEntry

_LOGGER = logging.getLogger(__name__)

type Device = Enclosure | System | PlantSite
type MetricFetcher = Callable[[DemeterClient, int], Awaitable[Reading | None]]

# Which API method pulls the latest reading of each metric, per device kind.
_METRIC_FETCHERS: dict[DeviceKind, dict[Metric, MetricFetcher]] = {
    DeviceKind.ENCLOSURE: {
        Metric.AIR_TEMPERATURE: DemeterClient.latest_air_temperature,
        Metric.AIR_HUMIDITY: DemeterClient.latest_air_humidity,
    },
    DeviceKind.SYSTEM: {
        Metric.PH: DemeterClient.latest_ph,
        Metric.EC: DemeterClient.latest_ec,
        Metric.WATER_TEMPERATURE: DemeterClient.latest_water_temperature,
    },
    DeviceKind.PLANT_SITE: {
        Metric.LUX: DemeterClient.latest_lux,
    },
}


@dataclass
class DemeterData:
    """Snapshot of the demeter topology and the latest metric readings."""

    enclosures: dict[int, Enclosure] = field(default_factory=dict)
    systems: dict[int, System] = field(default_factory=dict)
    flows: dict[int, Flow] = field(default_factory=dict)
    plant_sites: dict[int, PlantSite] = field(default_factory=dict)
    readings: dict[str, Reading] = field(default_factory=dict)

    def devices(self, kind: DeviceKind) -> list[Device]:
        """Return all discovered devices of the given kind."""
        match kind:
            case DeviceKind.ENCLOSURE:
                return list(self.enclosures.values())
            case DeviceKind.SYSTEM:
                return list(self.systems.values())
            case DeviceKind.PLANT_SITE:
                return list(self.plant_sites.values())


class DemeterCoordinator(DataUpdateCoordinator[DemeterData]):
    """Coordinate polling of the demeter API."""

    def __init__(
        self, hass: HomeAssistant, client: DemeterClient, entry: DemeterConfigEntry
    ) -> None:
        """Initialize the coordinator."""
        super().__init__(
            hass,
            _LOGGER,
            config_entry=entry,
            name=DOMAIN,
            update_interval=UPDATE_INTERVAL,
        )
        self.client = client

    async def _async_update_data(self) -> DemeterData:
        """Fetch the topology, then the latest reading of every metric."""
        try:
            enclosures = await self.client.list_enclosures()
            systems = await self.client.list_systems()
            flows = await self.client.list_flows()
            plant_sites = await self.client.list_plant_sites()
        except DemeterError as err:
            raise UpdateFailed(f"error communicating with demeter: {err}") from err

        data = DemeterData(
            enclosures={device.id: device for device in enclosures},
            systems={device.id: device for device in systems},
            flows={device.id: device for device in flows},
            plant_sites={device.id: device for device in plant_sites},
        )

        jobs: list[tuple[str, MetricFetcher, int]] = []
        for kind, metrics in _METRIC_FETCHERS.items():
            for device in data.devices(kind):
                for metric, fetcher in metrics.items():
                    jobs.append(
                        (reading_key(kind, device.id, metric), fetcher, device.id)
                    )

        results = await asyncio.gather(
            *(fetcher(self.client, device_id) for _, fetcher, device_id in jobs),
            return_exceptions=True,
        )

        for (key, _, _), result in zip(jobs, results, strict=True):
            if isinstance(result, asyncio.CancelledError):
                raise result
            if isinstance(result, DemeterError):
                # A single metric failing (e.g. a device deleted mid-update)
                # should not take down the whole poll; its sensor shows
                # unknown while the rest stay fresh.
                _LOGGER.warning("failed to fetch %s from demeter: %s", key, result)
                continue
            if isinstance(result, BaseException):
                raise UpdateFailed(
                    f"unexpected error fetching {key} from demeter: {result!r}"
                ) from result
            if result is not None:
                data.readings[key] = result

        return data
