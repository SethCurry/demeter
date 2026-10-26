"""A tiny mock of the demeter HTTP API for the Home Assistant test harness.

The integration talks to the endpoints served by ``internal/httpd``. Running
a real demeter server in Docker is awkward (it needs a seeded SQLite
database and the embedded UI assets), so this container serves a fixed
topology with deterministic readings instead. The response shapes match the
sqlc-generated Go structs: PascalCase fields and ``sql.NullTime`` timestamps
as ``{"Time": "...", "Valid": true}``.
"""

from __future__ import annotations

import json
import os
from datetime import UTC, datetime
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import urlparse

DEFAULT_PORT = 8249


def _timestamp() -> dict[str, object]:
    """Return a freshly generated ``sql.NullTime``-shaped timestamp."""
    return {"Time": datetime.now(UTC).isoformat(), "Valid": True}


def _reading(value_field: str, value: float) -> list[dict[str, object]]:
    """Return a newest-first reading list with one entry."""
    return [{value_field: value, "Timestamp": _timestamp()}]


def _responses() -> dict[str, object]:
    """Build the response bodies for every endpoint the integration uses."""
    return {
        "/api/enclosures": [{"ID": 1, "Name": "Test Enclosure"}],
        "/api/systems": [{"ID": 1, "Name": "Test System", "EnclosureID": 1}],
        "/api/flows": [{"ID": 1, "Name": "Test Flow", "SystemID": 1}],
        "/api/plant-sites": [{"ID": 1, "FlowID": 1, "X": 1, "Y": 2, "Z": 3}],
        "/api/enclosures/1/air-temperature": _reading("TemperatureC", 22.5),
        "/api/enclosures/1/air-humidity": _reading("HumidityRh", 55.0),
        "/api/systems/1/ph": _reading("Ph", 6.3),
        "/api/systems/1/ec": _reading("Ec", 1.8),
        "/api/systems/1/water-temperature": _reading("WaterTemperatureC", 19.5),
        "/api/plant-sites/1/lux": _reading("Lux", 12000),
    }


class DemeterMockHandler(BaseHTTPRequestHandler):
    """Serve the canned demeter responses."""

    server_version = "DemeterMock/1.0"

    def do_GET(self) -> None:  # noqa: N802 - BaseHTTPRequestHandler API
        """Handle a GET request."""
        path = urlparse(self.path).path.rstrip("/")
        body = _responses().get(path)
        if body is None:
            self.send_response(404)
            self.end_headers()
            return

        raw = json.dumps(body).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)

    def log_message(self, format: str, *args: object) -> None:  # noqa: A002
        """Log requests to stderr in the usual combined-log format."""
        print(f"{self.address_string()} - {format % args}", flush=True)


def main() -> None:
    """Run the mock server."""
    port = int(os.environ.get("PORT", DEFAULT_PORT))
    server = ThreadingHTTPServer(("0.0.0.0", port), DemeterMockHandler)
    print(f"demeter mock listening on port {port}", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
