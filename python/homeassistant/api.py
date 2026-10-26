"""Async client for the demeter HTTP API (see internal/httpd).

The demeter server exposes an unauthenticated JSON API. Responses are
marshaled straight from the sqlc-generated Go structs, which carry no JSON
tags: field names are PascalCase (``ID``, ``Name``, ``TemperatureC``, ...)
and nullable columns are serialized as database/sql ``Null*`` objects,
e.g. timestamps arrive as ``{"Time": "2025-...", "Valid": true}``.
"""

from __future__ import annotations

import logging
from dataclasses import dataclass
from datetime import UTC, datetime
from json import JSONDecodeError
from typing import Any

from aiohttp import ClientError, ClientSession, ClientTimeout

_LOGGER = logging.getLogger(__name__)

REQUEST_TIMEOUT = ClientTimeout(total=10)


class DemeterError(Exception):
    """Base error for problems talking to the demeter API."""


class DemeterConnectionError(DemeterError):
    """Raised when the demeter server cannot be reached or returns an error."""


class DemeterNotFoundError(DemeterError):
    """Raised when the demeter server reports a resource does not exist."""


@dataclass(frozen=True)
class Reading:
    """A single metric reading with the time it was taken."""

    value: float
    timestamp: datetime | None


@dataclass(frozen=True)
class Enclosure:
    """A demeter enclosure (the space the plants are in)."""

    id: int
    name: str


@dataclass(frozen=True)
class System:
    """A demeter system (plants sharing a reservoir)."""

    id: int
    name: str
    enclosure_id: int


@dataclass(frozen=True)
class Flow:
    """A demeter flow (a single loop of water in a system)."""

    id: int
    name: str
    system_id: int


@dataclass(frozen=True)
class PlantSite:
    """A demeter plant site (a location within a flow)."""

    id: int
    flow_id: int
    x: int
    y: int
    z: int


def _null_time(payload: Any) -> datetime | None:
    """Extract a datetime from a serialized ``sql.NullTime`` object."""
    if not isinstance(payload, dict) or not payload.get("Valid"):
        return None
    raw = payload.get("Time")
    if not isinstance(raw, str):
        raise DemeterError(f"malformed timestamp in demeter response: {payload!r}")
    try:
        parsed = datetime.fromisoformat(raw)
    except ValueError as err:
        raise DemeterError(f"malformed timestamp in demeter response: {raw!r}") from err
    if parsed.tzinfo is None:
        return parsed.replace(tzinfo=UTC)
    return parsed


def _as_list(payload: Any, what: str) -> list[Any]:
    """Validate that a response is a JSON list.

    Empty Go slices marshal as JSON ``null`` rather than ``[]``, so that is
    treated as an empty list.
    """
    if payload is None:
        return []
    if not isinstance(payload, list):
        raise DemeterError(f"demeter response for {what} was malformed: {payload!r}")
    return payload


def _field(item: Any, field: str, what: str) -> Any:
    """Return a required field of a response object."""
    try:
        return item[field]
    except (KeyError, IndexError, TypeError) as err:
        raise DemeterError(
            f"demeter response for {what} was malformed: {item!r}"
        ) from err


def _reading_from_list(payload: Any, value_field: str, what: str) -> Reading | None:
    """Convert the newest item of a reading list into a Reading.

    The reading endpoints return newest-first lists, so the first element is
    the most recent reading.
    """
    items = _as_list(payload, what)
    if not items:
        return None
    item = items[0]
    return Reading(
        value=_field(item, value_field, what),
        timestamp=_null_time(item.get("Timestamp")),
    )


class DemeterClient:
    """Thin async client around the demeter HTTP API."""

    def __init__(self, session: ClientSession, host: str) -> None:
        """Initialize the client with the server base URL."""
        self._session = session
        self._host = host.rstrip("/")

    @property
    def host(self) -> str:
        """The base URL of the demeter server."""
        return self._host

    async def _get_json(self, path: str, params: dict[str, str] | None = None) -> Any:
        """GET a JSON document from the demeter API."""
        url = f"{self._host}{path}"
        try:
            async with self._session.get(
                url, params=params, timeout=REQUEST_TIMEOUT
            ) as resp:
                if resp.status == 404:
                    raise DemeterNotFoundError(f"GET {path} returned 404")
                if resp.status >= 400:
                    raise DemeterConnectionError(f"GET {path} returned {resp.status}")
                return await resp.json()
        except TimeoutError as err:
            raise DemeterConnectionError(f"GET {path} timed out") from err
        except ClientError as err:
            raise DemeterConnectionError(f"GET {path} failed: {err}") from err
        except JSONDecodeError as err:
            raise DemeterError(f"GET {path} returned malformed JSON: {err}") from err

    async def _latest_reading(
        self,
        path: str,
        value_field: str,
        what: str,
        params: dict[str, str] | None = None,
    ) -> Reading | None:
        """Fetch the newest reading for a metric, if any exists."""
        try:
            payload = await self._get_json(path, params)
        except DemeterNotFoundError:
            return None
        return _reading_from_list(payload, value_field, what)

    # Discovery: topology

    async def list_enclosures(self) -> list[Enclosure]:
        """Return all enclosures known to demeter."""
        payload = await self._get_json("/api/enclosures")
        return [
            Enclosure(
                id=_field(item, "ID", "enclosures"),
                name=_field(item, "Name", "enclosures"),
            )
            for item in _as_list(payload, "enclosures")
        ]

    async def list_systems(self) -> list[System]:
        """Return all systems known to demeter."""
        payload = await self._get_json("/api/systems")
        return [
            System(
                id=_field(item, "ID", "systems"),
                name=_field(item, "Name", "systems"),
                enclosure_id=_field(item, "EnclosureID", "systems"),
            )
            for item in _as_list(payload, "systems")
        ]

    async def list_flows(self) -> list[Flow]:
        """Return all flows known to demeter."""
        payload = await self._get_json("/api/flows")
        return [
            Flow(
                id=_field(item, "ID", "flows"),
                name=_field(item, "Name", "flows"),
                system_id=_field(item, "SystemID", "flows"),
            )
            for item in _as_list(payload, "flows")
        ]

    async def list_plant_sites(self) -> list[PlantSite]:
        """Return all plant sites known to demeter."""
        payload = await self._get_json("/api/plant-sites")
        return [
            PlantSite(
                id=_field(item, "ID", "plant sites"),
                flow_id=_field(item, "FlowID", "plant sites"),
                x=_field(item, "X", "plant sites"),
                y=_field(item, "Y", "plant sites"),
                z=_field(item, "Z", "plant sites"),
            )
            for item in _as_list(payload, "plant sites")
        ]

    async def latest_air_temperature(self, enclosure_id: int) -> Reading | None:
        """Return the most recent air temperature reading of an enclosure."""
        return await self._latest_reading(
            f"/api/enclosures/{enclosure_id}/air-temperature",
            "TemperatureC",
            "enclosure air temperature",
            params={"limit": "1"},
        )

    async def latest_air_humidity(self, enclosure_id: int) -> Reading | None:
        """Return the most recent air humidity reading of an enclosure."""
        return await self._latest_reading(
            f"/api/enclosures/{enclosure_id}/air-humidity",
            "HumidityRh",
            "enclosure air humidity",
            params={"limit": "1"},
        )

    async def latest_ph(self, system_id: int) -> Reading | None:
        """Return the most recent pH reading of a system."""
        return await self._latest_reading(
            f"/api/systems/{system_id}/ph",
            "Ph",
            "system pH",
            params={"limit": "1"},
        )

    async def latest_ec(self, system_id: int) -> Reading | None:
        """Return the most recent EC reading of a system."""
        return await self._latest_reading(
            f"/api/systems/{system_id}/ec",
            "Ec",
            "system EC",
            params={"limit": "1"},
        )

    async def latest_water_temperature(self, system_id: int) -> Reading | None:
        """Return the most recent water temperature reading of a system."""
        return await self._latest_reading(
            f"/api/systems/{system_id}/water-temperature",
            "WaterTemperatureC",
            "system water temperature",
            params={"limit": "1"},
        )

    async def latest_lux(self, plant_site_id: int) -> Reading | None:
        """Return the most recent lux reading of a plant site."""
        return await self._latest_reading(
            f"/api/plant-sites/{plant_site_id}/lux",
            "Lux",
            "plant site lux",
        )
