"""The demeter integration.

Discovers enclosures, systems and plant sites from a demeter server and
exposes their metrics as Home Assistant sensor entities.
"""

from __future__ import annotations

from dataclasses import dataclass

from homeassistant.config_entries import ConfigEntry
from homeassistant.const import CONF_HOST, Platform
from homeassistant.core import HomeAssistant
from homeassistant.helpers.aiohttp_client import async_get_clientsession

from .api import DemeterClient
from .coordinator import DemeterCoordinator

PLATFORMS: list[Platform] = [Platform.SENSOR]


@dataclass
class DemeterRuntimeData:
    """Runtime data stored on a demeter config entry."""

    client: DemeterClient
    coordinator: DemeterCoordinator


type DemeterConfigEntry = ConfigEntry[DemeterRuntimeData]


async def async_setup_entry(hass: HomeAssistant, entry: DemeterConfigEntry) -> bool:
    """Set up demeter from a config entry."""
    client = DemeterClient(async_get_clientsession(hass), entry.data[CONF_HOST])
    coordinator = DemeterCoordinator(hass, client, entry)

    # The first refresh discovers the topology; a failure here makes the
    # config entry retry with ConfigEntryNotReady.
    await coordinator.async_config_entry_first_refresh()

    entry.runtime_data = DemeterRuntimeData(client=client, coordinator=coordinator)

    await hass.config_entries.async_forward_entry_setups(entry, PLATFORMS)
    return True


async def async_unload_entry(hass: HomeAssistant, entry: DemeterConfigEntry) -> bool:
    """Unload a demeter config entry."""
    return await hass.config_entries.async_unload_platforms(entry, PLATFORMS)
