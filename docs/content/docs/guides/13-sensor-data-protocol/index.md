---
title: "Sensor Data Protocol"
summary: "The exact WebSocket wire format for scalar readings and camera frames, including message types, integer scaling, and validation rules."
date: 2023-09-07T16:13:18+02:00
lastmod: 2026-11-03T00:00:00+00:00
draft: false
weight: 13
toc: true
---

Demeter's sensor protocol is intentionally small. Scalar readings are short
text messages over a WebSocket; camera frames are binary messages on the same
endpoint. There is no JSON envelope and no acknowledgement.

<figure>
  <img src="/images/sample.jpg" alt="WebSocket sensor message flow" width="1200" height="675" loading="lazy">
  <figcaption>Sensors connect, send text readings, and optionally send one binary image. Replace this placeholder with a diagram or screenshot.</figcaption>
</figure>

## Endpoint

```text
ws://<server>:8249/api/websocket/sensors
```

The endpoint is **unauthenticated**. Anyone who can reach the port can submit
readings or images. Restrict network access at the firewall or reverse proxy
if the server is not on a trusted network.

Each WebSocket connection may send:

- Any number of text readings, then
- At most one binary image.

After a binary image, the server saves the image and stops reading the
connection. It does not send an acknowledgement and does not explicitly close
the socket. Use a fresh connection for any further readings rather than
waiting for a close frame.

## Text readings

A text frame is a colon-delimited ASCII string:

```text
typeID:targetID:value
```

All three parts are required. The server parses `typeID` and `targetID` as
integers. For measurement types it parses `value` as a signed 64-bit integer.

```text
4:1:6420
```

This means: message type `4` (system pH), target system `1`, value `6420`,
which is stored as pH `6.420`.

### Rules

- Use a colon (`:`) separator. No spaces, JSON, or units.
- For measurement types, `value` must be an integer. Send `6420`, not `6.42`.
  A decimal point stops processing for that connection.
- Values can be negative where the metric allows it, such as temperature.
- Extra colon-separated fields are ignored; only the first three are read.
- An unknown `typeID` is logged and ignored; the connection stays open.
- A malformed message stops processing for that connection. The server does not
  explicitly close the socket, so the client may look connected while nothing
  is read. Close and reconnect from the client.
- The server does not send acknowledgements or errors back.

### Message types

| ID | Constant | Target | Value encoding | Stored field |
| --- | --- | --- | --- | --- |
| `0` | `Unknown` | — | — | Reserved. Unknown types are ignored. |
| `1` | `Initialize` | parsed, then ignored | ignored | No-op; the server logs the message. |
| `2` | `EnclosureAirTemperature` | enclosure ID | °C × 1000 | `enclosure_air_temperature.temperature_c` |
| `3` | `EnclosureHumidity` | enclosure ID | % RH × 1000 | `enclosure_air_humidity.humidity_rh` |
| `4` | `SystemPH` | system ID | pH × 1000 | `system_ph.ph` |
| `5` | `SystemEC` | system ID | EC × 1000 | `system_ec.ec` |
| `6` | `SystemOxygen` | system ID | value × 1000 | `system_oxygen.oxygen` |
| `7` | `SystemWaterTemperature` | system ID | °C × 1000 | `system_water_temperature.water_temperature_c` |
| `8` | `SystemWaterProximity` | system ID | metres × 1000 | `system_water_proximity.proximity_m` |
| `9` | `FlowRate` | flow ID | m/s × 1000 | `flow_rate.rate_mps` |
| `10` | `PlantSiteLux` | plant site ID | lux (no scaling) | `plant_site_lux.lux` |

Only types `2`, `3`, `4`, `5`, `7`, and `10` currently have REST read
endpoints or UI charts. Types `6`, `8`, and `9` are stored but not yet
readable through the API.

### Scaling examples

| Reading | Frame |
| --- | --- |
| Enclosure 1 at 24.5 °C | `2:1:24500` |
| Enclosure 1 at 61.2 % RH | `3:1:61200` |
| System 1 at pH 6.42 | `4:1:6420` |
| System 1 at EC 1.85 | `5:1:1850` |
| System 1 water at 21.75 °C | `7:1:21750` |
| Flow 3 at 1.25 m/s | `9:3:1250` |
| Plant site 56 at 430 lux | `10:56:430` |

Negative values are valid for temperatures:

```text
7:1:-2500
```

This records −2.5 °C.

## Initialize

The stock firmware does **not** send an initialize frame. The
`simulate-sensor` command sends one when it connects:

```text
1:0:0
```

The target field must still parse as an integer, but the target and third field
are otherwise ignored. `Initialize` currently carries no fields and only
writes a log line. It is safe to omit.

## Binary camera frames

A binary WebSocket message is a JPEG with a 12-byte MAC prefix:

```text
offset 0                  offset 12
+-------------------------+----------------------------+
| MAC as 12 ASCII bytes   | JPEG payload               |
+-------------------------+----------------------------+
```

Format the MAC as 12 lowercase hex characters, for example:

```text
a1b2c3d4e5f6
```

The server does not validate or normalise this prefix. It uses the 12 bytes
verbatim as the directory name, so a different length or letter case creates
a different directory. Stick to 12 lowercase hex characters.

For example, a frame from that device is stored under:

```text
/srv/demeter/timelapse/a1b2c3d4e5f6/<timestamp>.jpeg
```

The remainder of the message is written to disk verbatim as a `.jpeg` file.
There is no length prefix, checksum, or content-type header.

### Sending a binary frame

1. Send all text readings first.
2. Send one binary message whose first 12 bytes are the MAC and whose
   remaining bytes are the JPEG.
3. Do not wait for an acknowledgement. The server stops reading after saving
   the image but does not explicitly close the socket. Open a fresh connection
   for more readings.

The stock firmware sends the MAC as one frame, the JPEG in 512-byte fragments,
and a final empty frame so the server receives a single complete WebSocket
message. A client library that supports binary messages can send the combined
buffer in one call.

Do not send a binary frame with fewer than 12 bytes. The server reads the
first 12 bytes as the MAC without checking the message length first; a short
frame can panic the handler and leave the connection unserviced. Always send a
complete prefix and a JPEG.

## Multiple readings per connection

A device can publish several metrics before disconnecting:

```text
2:1:24500
3:1:61200
7:1:21750
10:56:430
```

Each text frame is processed independently and inserted immediately. A
database error is logged and the connection stays open, but that reading is
lost.

The server does not enforce SQLite foreign keys. A reading for a target ID that
does not exist is inserted successfully as an orphan row; it simply does not
appear under any real enclosure, system, flow, or plant site. Check target IDs
carefully. If a chart stays empty, confirm both the ID and the message type.

## Timestamps

The protocol has no timestamp field. The server stamps each reading with the
current time when it is inserted. For accurate history, keep the server clock
synchronised with NTP. A device with a wrong clock cannot correct the server's
timestamps.

## Minimal client example

The following Python client sends one pH reading and one image. It assumes the
`websockets` package.

```python
import asyncio
import pathlib
import websockets

async def main():
    async with websockets.connect(
        "ws://192.168.1.20:8249/api/websocket/sensors"
    ) as ws:
        await ws.send("4:1:6420")          # pH 6.420 for system 1
        await ws.send("7:1:21750")         # 21.750 °C for system 1

        mac = b"a1b2c3d4e5f6"
        jpeg = pathlib.Path("frame.jpg").read_bytes()
        await ws.send(mac + jpeg)

asyncio.run(main())
```

The server stops reading the socket after the image. It does not send an
acknowledgement or a close frame, so close the connection yourself and
reconnect before sending more readings.

## Home Assistant and other clients

The REST API used by Home Assistant is separate from this WebSocket protocol.
Home Assistant does not publish sensor readings; it reads them over HTTP. See
[Connecting to Home Assistant](../connecting-to-home-assistant).

For the HTTP side, see [REST API](../rest-api).
