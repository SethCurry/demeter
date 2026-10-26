#!/usr/bin/env python3
"""End-to-end smoke test for the demeter Home Assistant integration.

The test drives a real Home Assistant container through its REST API:

1. wait for Home Assistant to finish starting;
2. complete onboarding (or log in if it has already been done);
3. replace any existing demeter config entry with a fresh one pointing at
   the mock demeter API started by ``docker-compose.yml``;
4. wait for the config entry to load and verify every expected sensor state,
   unit and ``reading_timestamp`` attribute.

The script only uses the Python standard library, so it can run on the host
without installing anything.
"""

from __future__ import annotations

import json
import math
import os
import sys
import time
from datetime import datetime
from typing import Any
from urllib.error import HTTPError, URLError
from urllib.parse import urlencode
from urllib.request import Request, urlopen

HA_URL = os.environ.get("HA_URL", "http://localhost:8123").rstrip("/")
DEMETER_URL = os.environ.get("DEMETER_URL", "http://demeter-mock:8249").rstrip("/")
HA_USERNAME = os.environ.get("HA_USERNAME", "demeter-test")
HA_PASSWORD = os.environ.get("HA_PASSWORD", "demeter-test")
HA_CLIENT_ID = os.environ.get("HA_CLIENT_ID", f"{HA_URL}/")

READY_TIMEOUT = float(os.environ.get("HA_READY_TIMEOUT", "180"))
SETUP_TIMEOUT = float(os.environ.get("HA_SETUP_TIMEOUT", "60"))
STATE_TIMEOUT = float(os.environ.get("HA_STATE_TIMEOUT", "90"))
REQUEST_TIMEOUT = 30.0

# Entity id -> (expected value, expected unit). The names come from the mock
# topology in mock_demeter.py.
EXPECTED_SENSORS: dict[str, tuple[float, str | None]] = {
    "sensor.test_enclosure_air_temperature": (22.5, "°C"),
    "sensor.test_enclosure_air_humidity": (55.0, "%"),
    "sensor.test_system_ph": (6.3, None),
    "sensor.test_system_ec": (1.8, "μS/cm"),
    "sensor.test_system_water_temperature": (19.5, "°C"),
    "sensor.test_flow_site_1_2_3_illuminance": (12000, "lx"),
}


class SmokeTestError(RuntimeError):
    """Raised when the test cannot proceed or an expectation fails."""


def _decode(raw: bytes) -> Any:
    """Decode a JSON response body, tolerating empty or non-JSON bodies."""
    if not raw:
        return None
    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        return raw.decode(errors="replace")


class HomeAssistant:
    """Minimal REST client for the endpoints used by the smoke test."""

    def __init__(self, base_url: str) -> None:
        self.base_url = base_url
        self.access_token: str | None = None

    def request(
        self,
        method: str,
        path: str,
        payload: dict[str, Any] | None = None,
        *,
        token: str | None = None,
        form: bool = False,
    ) -> tuple[int, Any]:
        """Perform a request and return ``(status, decoded_body)``."""
        headers = {"Accept": "application/json"}
        data = None
        if payload is not None:
            if form:
                data = urlencode(payload).encode()
                headers["Content-Type"] = "application/x-www-form-urlencoded"
            else:
                data = json.dumps(payload).encode()
                headers["Content-Type"] = "application/json"
        if token is not None:
            headers["Authorization"] = f"Bearer {token}"

        request = Request(
            f"{self.base_url}{path}", data=data, headers=headers, method=method
        )
        try:
            with urlopen(request, timeout=REQUEST_TIMEOUT) as response:
                return response.status, _decode(response.read())
        except HTTPError as err:
            return err.code, _decode(err.read())

    def api(
        self,
        method: str,
        path: str,
        payload: dict[str, Any] | None = None,
    ) -> Any:
        """Perform an authenticated API request, raising on failure."""
        if self.access_token is None:
            raise SmokeTestError("not authenticated")
        status, body = self.request(method, path, payload, token=self.access_token)
        if status >= 400:
            raise SmokeTestError(
                f"{method} {path} failed with HTTP {status}: {body!r}"
            )
        return body

    def wait_until_ready(self) -> None:
        """Wait until the Home Assistant HTTP API is accepting requests."""
        deadline = time.monotonic() + READY_TIMEOUT
        last_error = "no response yet"
        while time.monotonic() < deadline:
            try:
                # /auth/providers returns 200 once onboarding is complete.
                # While onboarding is still in progress it returns 400, so in
                # that case wait for /api/onboarding to be registered.
                status, body = self.request("GET", "/auth/providers")
                if (
                    status == 400
                    and isinstance(body, dict)
                    and body.get("code") == "onboarding_required"
                ):
                    status, body = self.request("GET", "/api/onboarding")
                    if status == 200:
                        print(f"Home Assistant is ready at {self.base_url}")
                        return
            except (URLError, TimeoutError, OSError) as err:
                last_error = str(err)
            else:
                if status == 200:
                    print(f"Home Assistant is ready at {self.base_url}")
                    return
                last_error = f"HTTP {status}: {body!r}"
            time.sleep(2)

        raise SmokeTestError(
            f"Home Assistant did not become ready within {READY_TIMEOUT:.0f}s "
            f"({last_error})"
        )

    def _onboarding_status(self) -> dict[str, bool]:
        """Return the onboarding steps mapped to their done flag."""
        status, body = self.request("GET", "/api/onboarding")
        if status == 404:
            # The onboarding views are not registered once onboarding has
            # completed, so a 404 means every step is done.
            return {
                "user": True,
                "core_config": True,
                "analytics": True,
                "integration": True,
            }
        if status != 200 or not isinstance(body, list):
            raise SmokeTestError(
                f"failed to read onboarding status: HTTP {status}: {body!r}"
            )
        return {item["step"]: item["done"] for item in body}

    def _login_code(self) -> str:
        """Log in with the configured credentials and return an auth code."""
        status, flow = self.request(
            "POST",
            "/auth/login_flow",
            {
                "client_id": HA_CLIENT_ID,
                "handler": ["homeassistant", None],
                "redirect_uri": HA_CLIENT_ID,
            },
        )
        if status != 200 or not isinstance(flow, dict) or "flow_id" not in flow:
            raise SmokeTestError(f"could not start login flow: HTTP {status}: {flow!r}")

        status, result = self.request(
            "POST",
            f"/auth/login_flow/{flow['flow_id']}",
            {
                "client_id": HA_CLIENT_ID,
                "username": HA_USERNAME,
                "password": HA_PASSWORD,
            },
        )
        if (
            status != 200
            or not isinstance(result, dict)
            or result.get("type") != "create_entry"
        ):
            raise SmokeTestError(f"login failed: HTTP {status}: {result!r}")
        return result["result"]

    def _exchange_code(self, auth_code: str) -> str:
        """Exchange an auth code for an access token."""
        status, token = self.request(
            "POST",
            "/auth/token",
            {
                "grant_type": "authorization_code",
                "code": auth_code,
                "client_id": HA_CLIENT_ID,
            },
            form=True,
        )
        if status != 200 or not isinstance(token, dict):
            raise SmokeTestError(
                f"could not exchange auth code: HTTP {status}: {token!r}"
            )
        return token["access_token"]

    def authenticate(self) -> None:
        """Create the test user or log in, then finish onboarding."""
        steps = self._onboarding_status()

        if steps.get("user"):
            auth_code = self._login_code()
        else:
            status, body = self.request(
                "POST",
                "/api/onboarding/users",
                {
                    "name": "Demeter Test",
                    "username": HA_USERNAME,
                    "password": HA_PASSWORD,
                    "client_id": HA_CLIENT_ID,
                    "language": "en",
                },
            )
            if status != 200 or not isinstance(body, dict):
                raise SmokeTestError(
                    f"could not create onboarding user: HTTP {status}: {body!r}"
                )
            auth_code = body["auth_code"]

        token = self._exchange_code(auth_code)

        # Finish the remaining onboarding steps so the UI is usable after a
        # manual run. The integration step requires a valid client id and
        # redirect URI; the local Home Assistant URL is accepted.
        steps = self._onboarding_status()
        if not steps.get("core_config"):
            status, body = self.request(
                "POST", "/api/onboarding/core_config", {}, token=token
            )
            if status not in (200, 403):
                raise SmokeTestError(
                    f"core_config onboarding failed: HTTP {status}: {body!r}"
                )
        if not steps.get("analytics"):
            status, body = self.request(
                "POST", "/api/onboarding/analytics", token=token
            )
            if status not in (200, 403):
                raise SmokeTestError(
                    f"analytics onboarding failed: HTTP {status}: {body!r}"
                )
        if not steps.get("integration"):
            status, body = self.request(
                "POST",
                "/api/onboarding/integration",
                {"client_id": HA_CLIENT_ID, "redirect_uri": HA_CLIENT_ID},
                token=token,
            )
            if status not in (200, 403):
                raise SmokeTestError(
                    f"integration onboarding failed: HTTP {status}: {body!r}"
                )

        print("Authenticated with Home Assistant")
        self.access_token = token

    def replace_demeter_entry(self, host: str) -> str:
        """Delete existing demeter entries and create a fresh one."""
        entries = self.api("GET", "/api/config/config_entries/entry")
        removed = 0
        for entry in entries:
            if entry["domain"] == "demeter":
                self.api(
                    "DELETE",
                    f"/api/config/config_entries/entry/{entry['entry_id']}",
                )
                removed += 1

        if removed:
            suffix = "y" if removed == 1 else "ies"
            print(f"Removed {removed} existing demeter config entr{suffix}")
            # Give Home Assistant a moment to tear down the old entry before
            # re-using the same unique id.
            time.sleep(1)

        status, flow = self.request(
            "POST",
            "/api/config/config_entries/flow",
            {"handler": "demeter"},
            token=self.access_token,
        )
        if status != 200 or not isinstance(flow, dict):
            raise SmokeTestError(
                f"could not start demeter config flow: HTTP {status}: {flow!r}"
            )
        if flow.get("type") != "form":
            raise SmokeTestError(f"unexpected demeter config flow result: {flow!r}")

        status, result = self.request(
            "POST",
            f"/api/config/config_entries/flow/{flow['flow_id']}",
            {"host": host},
            token=self.access_token,
        )
        if status != 200 or not isinstance(result, dict):
            raise SmokeTestError(
                f"could not create demeter config entry: HTTP {status}: {result!r}"
            )
        if result.get("type") != "create_entry":
            raise SmokeTestError(f"demeter config entry was not created: {result!r}")

        entry_id = result["result"]["entry_id"]
        print(f"Created demeter config entry {entry_id} for {host}")
        return entry_id

    def wait_for_entry_loaded(self, entry_id: str) -> dict[str, Any]:
        """Wait for a config entry to reach the ``loaded`` state."""
        deadline = time.monotonic() + SETUP_TIMEOUT
        last_entry: dict[str, Any] | None = None
        while time.monotonic() < deadline:
            entries = self.api("GET", "/api/config/config_entries/entry")
            for entry in entries:
                if entry["entry_id"] == entry_id:
                    last_entry = entry
                    if entry["state"] == "loaded":
                        print(f"Config entry {entry_id} is loaded")
                        return entry
            time.sleep(1)

        raise SmokeTestError(
            f"config entry {entry_id} did not load within {SETUP_TIMEOUT:.0f}s "
            f"(last state: {last_entry!r})"
        )

    def wait_for_sensors(self) -> dict[str, dict[str, Any]]:
        """Wait for all expected sensor entities and validate their values."""
        deadline = time.monotonic() + STATE_TIMEOUT
        last_errors: list[str] = []

        while time.monotonic() < deadline:
            states = self.api("GET", "/api/states")
            by_id = {state["entity_id"]: state for state in states}
            errors: list[str] = []

            for entity_id, (expected_value, expected_unit) in EXPECTED_SENSORS.items():
                state = by_id.get(entity_id)
                if state is None:
                    errors.append(f"{entity_id}: entity is missing")
                    continue

                try:
                    actual_value = float(state["state"])
                except (KeyError, TypeError, ValueError):
                    errors.append(
                        f"{entity_id}: state is not numeric ({state.get('state')!r})"
                    )
                    continue

                if not math.isclose(actual_value, expected_value, rel_tol=1e-6):
                    errors.append(
                        f"{entity_id}: value {actual_value!r} != {expected_value!r}"
                    )

                attributes = state.get("attributes", {})
                actual_unit = attributes.get("unit_of_measurement")
                if actual_unit != expected_unit:
                    errors.append(
                        f"{entity_id}: unit {actual_unit!r} != {expected_unit!r}"
                    )

                timestamp = attributes.get("reading_timestamp")
                if not isinstance(timestamp, str):
                    errors.append(f"{entity_id}: reading_timestamp is missing")
                    continue
                try:
                    datetime.fromisoformat(timestamp)
                except ValueError:
                    errors.append(
                        f"{entity_id}: reading_timestamp is malformed ({timestamp!r})"
                    )

            if not errors:
                return {entity_id: by_id[entity_id] for entity_id in EXPECTED_SENSORS}

            last_errors = errors
            time.sleep(2)

        raise SmokeTestError(
            f"sensors did not become ready within {STATE_TIMEOUT:.0f}s:\n  "
            + "\n  ".join(last_errors)
        )


def main() -> int:
    """Run the smoke test."""
    hass = HomeAssistant(HA_URL)
    hass.wait_until_ready()
    hass.authenticate()

    entry_id = hass.replace_demeter_entry(DEMETER_URL)
    hass.wait_for_entry_loaded(entry_id)
    states = hass.wait_for_sensors()

    print("\nVerified sensors:")
    for entity_id, state in states.items():
        attributes = state["attributes"]
        unit = attributes.get("unit_of_measurement") or ""
        print(
            f"  {entity_id} = {state['state']} {unit}"
            f" (reading_timestamp={attributes['reading_timestamp']})"
        )

    print("\nSmoke test passed.")
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except SmokeTestError as err:
        print(f"\nSmoke test failed: {err}", file=sys.stderr)
        sys.exit(1)
    except KeyboardInterrupt:
        sys.exit(130)
