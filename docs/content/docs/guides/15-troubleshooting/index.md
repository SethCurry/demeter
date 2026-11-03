---
title: "Troubleshooting"
summary: "Diagnose startup, networking, sensor, chart, timelapse, and Home Assistant problems with concrete checks and log messages."
date: 2023-09-07T16:13:18+02:00
lastmod: 2026-11-03T00:00:00+00:00
draft: false
weight: 15
toc: true
---

Most Demeter problems fall into one of five buckets: the server is not
running, the sensor cannot reach it, the target ID is wrong, the database is
not being written, or the timelapse directory is not persistent. Work through
the checks below in order.

<figure>
  <img src="/images/sample.jpg" alt="Server and sensor logs during troubleshooting" width="1200" height="675" loading="lazy">
  <figcaption>Server logs show received readings and database errors. Replace this placeholder with a screenshot.</figcaption>
</figure>

## First checks

### Is the server listening?

```sh
curl -i http://localhost:8249/api/enclosures
```

A healthy fresh server returns `200` and JSON `null` (not `[]`) for an empty
list. If the connection is refused:

- Confirm `./demeter server` is still running.
- Check the address. The default is `:8249`; `--address` overrides it.
- If connecting from another machine, bind to a reachable interface such as
  `--address 0.0.0.0:8249`.
- Check the host firewall. Port 8249 must be open to the sensor VLAN and to
  Home Assistant.

### Is the web panel served?

Open `http://localhost:8249` in a browser. The server serves the embedded UI at
`/`; direct requests to client-side routes such as `/enclosures` return a plain
`404 page not found`. Start at the root and use the sidebar. A blank page or
stale UI usually means the binary was built without running the UI embed step:

```sh
task build
```

`task build` runs `task embed`, which builds `web/`, copies the output into
`internal/embedded/ui`, and copies `sql/schema.sql`. A plain
`go build ./cmd/demeter` alone does not refresh the embedded UI or schema.

### Is the database writable?

The server opens `./demeter.sqlite` relative to its **working directory** and
attempts to apply the embedded schema on startup, ignoring any error. If that
directory is not writable, reads may work while every insert fails.

```sh
ls -l demeter.sqlite
sqlite3 demeter.sqlite '.tables'
```

A fresh database should contain tables such as `enclosure`, `system`, `flow`,
`plant_site`, `plant`, `system_ph`, and `enclosure_air_temperature`. If the
file is missing, check where you started the process.

## Server log messages

Logs go to stderr in a human-readable format. Useful lines include:

| Log message | Meaning |
| --- | --- |
| Gin route output | Startup; the HTTP server is registering routes and listening. |
| `received websocket message` | A sensor frame arrived. |
| `received enclosure air temperature message` | A valid scalar reading was parsed. |
| `failed to record ...` | The reading was parsed but the database insert failed, for example because the database is unwritable or the schema is missing. A missing target ID does **not** cause this. |
| `invalid message type ID`, `invalid target ID`, `invalid measurement value` | Malformed sensor frame; the server stops processing that connection. |
| `unknown message type` | Parsed frame with an unrecognised type ID; ignored. |
| `received timelapse image message` | A binary camera frame arrived. |
| `saved timelapse image` | The server attempted to write the frame. It does not check the write result, so verify the file on disk. |
| `failed to take picture` | Firmware-side camera failure. |

If you see no `received websocket message` at all, the problem is before the
database: networking, URL, or the device never connecting.

## Sensor connects but readings do not appear

### Check the target ID

Every reading should target an existing row:

- Enclosure temperature/humidity &rarr; `enclosure.ID`
- System pH/EC/oxygen/water temperature/proximity &rarr; `system.ID`
- Flow rate &rarr; `flow.ID`
- Lux &rarr; `plant_site.ID`

The current server does not enforce SQLite foreign keys. A wrong ID does not
produce a database error; it is stored as an orphan reading that never appears
under a real record. There is no error log to find. Verify the ID directly:

```sh
curl http://localhost:8249/api/enclosures
curl http://localhost:8249/api/systems
curl http://localhost:8249/api/plant-sites
```

### Check the WebSocket URL

The path is exactly `/api/websocket/sensors`:

```text
ws://192.168.1.20:8249/api/websocket/sensors
```

Common mistakes:

- `http://` instead of `ws://`.
- Missing port, so the client tries port 80.
- Missing `/api/websocket/sensors`.
- `localhost`, which on the ESP32 means the ESP32 itself.

Test from another machine with a WebSocket client before blaming the firmware.

### Check the message format

Frames must be:

```text
typeID:targetID:value
```

with integer values. `4:1:6.42` is rejected; use `4:1:6420`. See
[Sensor Data Protocol](../sensor-data-protocol).

### Check the sensor interval

Scalar sensors publish every `SENSOR_INTERVAL_SECONDS` (15 seconds by
default). If you just flashed the device, wait at least one interval and
refresh the page.

## Charts are empty

What an empty chart means depends on the page.

**Enclosure air charts** use metric endpoints and the selected time window. The
default is the last 24 hours. A chart can be empty while readings exist outside
that window.

- Check the selected range.
- Confirm the **server** clock is correct. The protocol has no timestamps;
  the server stamps each reading on arrival, so the sensor clock does not
  affect chart placement. A server with the wrong time writes readings into
  the wrong window.
- Query the metric endpoint directly:

```sh
curl 'http://localhost:8249/api/enclosures/1/metrics/air-temperature?limit=24'
```

An empty array means no readings in the selected window. If raw readings exist
but metrics are empty, check `start`, `end`, and the server's time zone. Use
explicit UTC timestamps. If the chart has gaps, those intervals had no
readings. This is expected.

**System charts and latest-value cards** share the same raw readings, limited
to the 20 newest rows per metric. If the card says **No readings yet**, the
chart is empty too. Query the raw endpoint, which may return `null` when empty:

```sh
curl 'http://localhost:8249/api/systems/1/ph?limit=20'
```

**Plant Lux charts** use up to 100 raw lux readings for the plant site and
have no range control. Query `/api/plant-sites/{id}/lux` to confirm whether
rows exist.

## Latest-value cards are empty

System detail pages request raw readings and display the newest value; when
there are no readings they show **No readings yet**. The sensor may not be
reporting, or the target ID may be wrong. Use the raw endpoint to confirm.
Enclosure pages show air-temperature and air-humidity charts but no separate
latest-value cards, and plant pages show a lux chart rather than a latest-value
card.

## The sensor appears to disconnect or stops reporting

After a binary image, the server saves the frame and stops reading that
connection, but it does not explicitly close the socket. The stock firmware
uses one connection for all sensors and may therefore stop delivering readings
after its first image unless the client reconnects; it may not detect that
anything is wrong because the server leaves the socket open. If scalar readings
stop after a camera frame, restart or reconnect the device, or move the camera
to a separate connection.

If a scalar-only device reconnects constantly, look for malformed frames in the
server log. A bad value stops processing for that connection, but the server
may not close the socket, so the device can appear connected while its readings
are ignored. Also check WiFi signal strength and power supply stability.

## I2C sensors do not start

- `I2C_DATA_GPIO` must be non-zero. The firmware does not initialise the bus
  otherwise.
- `I2C_CLOCK_GPIO` must be a valid output pin.
- The SHT4x starts only when `SHT_ENCLOSURE_ID` is non-zero.
- The BH1750 starts only when `BH_PLANT_SITE_ID` is non-zero.
- Check wiring: 3.3 V, ground, SCL, SDA.
- Check that the breakout's pull-ups are appropriate. Too many parallel
  pull-ups can also cause failures.
- Scan the bus with an I2C scanner sketch if you suspect wiring.

## DS18B20 reads nothing

- `DS18B20_GPIO` must be non-zero.
- A 4.7 kΩ pull-up from data to 3.3 V is required.
- The firmware uses the first device on the bus. If multiple probes are
  connected, only one is read.
- Long cables can corrupt 1-Wire traffic. Shorten the cable or improve
  grounding.

## Camera and timelapse problems

### No image cards

Check the server log for `received timelapse image message`. If it never
appears, the camera is not sending, or the server has already stopped reading
that connection after an earlier image.

- The firmware captures at most every five minutes.
- Camera capture needs PSRAM and adequate power.
- The stock firmware may stop receiving service after its first image because
  the server does not close and the client may not reconnect promptly.
- Look for `failed to take picture` or camera initialisation errors in the
  serial monitor.

### Image saved but playback fails

- Confirm `/srv/demeter/timelapse/<mac>/` contains `.jpeg` files.
- Confirm the server process can read them.
- Try the MJPEG URL directly in a browser.
- Corrupt frames are skipped with `failed to parse image`.

### Images vanish after a container restart

The dev Compose stack persists only the SQLite volume (`demeter-data`), not
`/srv/demeter/timelapse`. Add a bind mount or named volume for the timelapse
directory. See [Photos and Timelapse](../photos-and-timelapse#docker-warning).

## Home Assistant problems

### Cannot connect during setup

Home Assistant must reach Demeter from **its own** network context:

```sh
# From inside the Home Assistant container, if you have a shell:
curl http://host.docker.internal:8249/api/enclosures
```

Use `http://host.docker.internal:8249` when Home Assistant is containerised
and Demeter runs on the host. Use the Compose service name when both are on
the same network. See
[Connecting to Home Assistant](../connecting-to-home-assistant).

### Entities are unknown

The integration reports `unknown` when Demeter has no reading for that metric.
Create the entity in Demeter, attach a sensor, and wait for the next 60-second
poll. If only one metric is unknown, one reading endpoint is probably failing;
check the Demeter log.

### Entities do not update after a rename

Home Assistant fixes device names at creation. Remove and re-add the
integration, or rename the device in Home Assistant. Demeter renames do not
propagate.

## API problems

### `404` on an API route

Check the path and method against [REST API](../rest-api). Unknown paths return
a plain-text `404 page not found`; they do not fall through to the web panel.

### `409 Conflict` or an empty `500` on create

The name is probably already used. Names are global and unique per resource
type. Genus and species creation return `409` with a JSON error. Enclosure,
system, and flow creation currently return `500` with an empty body for the
same duplicate-name condition, because the handler checks the primary SQLite
constraint code while the driver reports an extended code.

Neither response means that a referenced parent is missing: foreign keys are
not enforced, so a bad parent ID is usually stored without an error. Choose a
different name, then verify that any parent ID you supplied points at the
record you intended.

### `400` on a metric query

- `start` and `end` must be RFC3339 timestamps.
- `start` must be before `end`.
- `limit` must be between 1 and 10000.

### Response fields look wrong

Responses use PascalCase and `Null*` wrapper objects:

```json
{ "ParentFlowID": { "Int64": 0, "Valid": false } }
```

See [REST API](../rest-api#response-field-casing).

## Starting over

For a disposable development database:

```sh
# Stop the server first.
rm demeter.sqlite
./demeter server
```

When the database file is missing, startup creates the schema. This deletes
all topology, readings, and notes. Back up the file first if you might need
it.

## Still stuck?

Collect:

- The exact server command and working directory.
- The full server log around the failure.
- The sensor's serial monitor output.
- The output of `curl -i http://<server>:8249/api/enclosures`.
- The result of the relevant raw and metric endpoints.

Then open an issue with those details. See
[Operating and Maintaining Demeter](../operating-and-maintaining-demeter) for log
and backup procedures.
