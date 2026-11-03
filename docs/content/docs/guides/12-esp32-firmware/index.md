---
title: "Building and Flashing the ESP32 Firmware"
summary: "Install ESP-IDF, configure WiFi, the server URL, GPIOs, and target IDs, then build, flash, and monitor the sensor node."
date: 2023-09-07T16:13:18+02:00
lastmod: 2026-11-03T00:00:00+00:00
draft: false
weight: 12
toc: true
---

The firmware in `esp-sensor/` is an ESP-IDF project. It connects to WiFi,
opens a WebSocket to Demeter, and publishes scalar readings and camera frames
on a schedule.

<figure>
  <img src="/images/sample.jpg" alt="ESP-IDF monitor showing sensor readings" width="1200" height="675" loading="lazy">
  <figcaption>A successful boot shows WiFi, WebSocket, and sensor log lines. Replace this placeholder with a screenshot.</figcaption>
</figure>

## Prerequisites

- **ESP-IDF 5.x** with the ESP32-S3 toolchain. Follow Espressif's
  [ESP-IDF getting-started guide](https://docs.espressif.com/projects/esp-idf/en/latest/esp32s3/get-started/)
  for your operating system.
- A USB cable that supports data, not just charging.
- Permission to access the serial device. On Linux, add your user to the
  `dialout` group and log in again.
- A running Demeter server reachable from the ESP32's network. The server must
  be listening on an address the device can reach, not only `127.0.0.1`.

Load the ESP-IDF environment in every new shell:

```sh
. "$IDF_PATH/export.sh"
```

If `IDF_PATH` is not set, source the `export.sh` from your ESP-IDF checkout.

## Configure the project

From the repository root:

```sh
cd esp-sensor
task config
```

This runs `idf.py menuconfig`. All project settings live under **Demeter**:

| Setting | Default | Purpose |
| --- | --- | --- |
| `SENSOR_INTERVAL_SECONDS` | `15` | Seconds between scalar sensor measurements. |
| `WIFI_SSID` | empty | WiFi network name. |
| `WIFI_PASSWORD` | empty | WiFi password. |
| `WEBSOCKET_URL` | empty | Full WebSocket URL, including path. |
| `DS18B20_GPIO` | `0` | 1-Wire data pin; `0` disables the probe. |
| `DS18B20_SYSTEM_ID` | `0` | Target system ID for water temperature. |
| `I2C_CLOCK_GPIO` | `0` | I2C SCL pin. |
| `I2C_DATA_GPIO` | `0` | I2C SDA pin; `0` disables the bus. |
| `SHT_ENCLOSURE_ID` | `0` | Target enclosure ID for air temperature/humidity; `0` disables. |
| `BH_PLANT_SITE_ID` | `0` | Target plant site ID for lux; `0` disables. |

Set the WebSocket URL to the server's `/api/websocket/sensors` endpoint:

```text
ws://192.168.1.20:8249/api/websocket/sensors
```

Use the server's LAN address, not `localhost`. If Demeter is behind TLS, use
`wss://` and a certificate the ESP32 trusts; the stock configuration does not
include a custom CA bundle.

### Target IDs must point to real records

Each target ID should refer to an existing record:

- `DS18B20_SYSTEM_ID` should be an existing system.
- `SHT_ENCLOSURE_ID` should be an existing enclosure.
- `BH_PLANT_SITE_ID` should be an existing plant site.

Create the records in the web panel first, then copy their numeric IDs. The
schema declares these columns as foreign keys, but the current server does not
enable SQLite foreign-key enforcement. A typo is therefore stored as an orphan
reading rather than rejected: it will not appear in any real entity's charts,
but it still occupies space in the database. Check the ID carefully.

### One sensor per ESP32

The stock firmware registers at most one of each sensor type. If you need two
reservoir probes, build a second node with a different `DS18B20_SYSTEM_ID`, or
extend the firmware to register multiple sensor descriptors.

## About tracked configuration

The repository contains an `sdkconfig` file with development settings. An
`sdkconfig` is machine- and network-specific: it can contain your WiFi name
and password, server address, and pin assignments. Do not commit your own
`sdkconfig` to a public repository.

Recommended practice:

1. Run `git rm --cached esp-sensor/sdkconfig` if it is tracked in your fork.
2. Add `esp-sensor/sdkconfig` and `esp-sensor/sdkconfig.old` to `.gitignore`.
3. Keep reusable, non-secret defaults in `sdkconfig.default`.
4. Put secrets in menuconfig locally, or generate `sdkconfig` in CI from
   protected variables.

The tracked development `sdkconfig` in this repository should be treated as an
example only. Do not reuse its credentials or network address.

## Build

```sh
cd esp-sensor
task build
```

Or call ESP-IDF directly:

```sh
idf.py build
```

The first build downloads managed components listed in
`main/idf_component.yml`:

- `espressif/esp_websocket_client`
- `espressif/ds18b20`
- `esp-idf-lib/sht4x`
- `esp-idf-lib/bh1750`
- `espressif/esp32-camera`

This requires network access. The build produces `build/esp-sensor.bin` and a
merged flash layout.

If the target is not already ESP32-S3, set it once:

```sh
idf.py set-target esp32s3
```

`set-target` regenerates `sdkconfig`; re-check your Demeter settings afterwards.

## Flash and monitor

Find the port:

```sh
ls /dev/ttyACM* /dev/ttyUSB* 2>/dev/null
```

Then flash and open the monitor:

```sh
idf.py -p /dev/ttyACM0 flash monitor
```

Exit the monitor with `Ctrl+]`. On some boards you must hold **BOOT** while
plugging in to enter the bootloader. If flashing fails with a permissions
error, fix the `dialout` group membership.

A healthy boot looks roughly like:

```text
I (1234) wifi: connected with YourNetwork, aid = 2, channel 6, BW20, bssid = ...
I (2345) demeter: connecting to ws://192.168.1.20:8249/api/websocket/sensors
I (2456) demeter: websocket connected
I (3000) demeter: publishing water temperature ...
```

Exact log lines vary by ESP-IDF version and enabled sensors.

## Runtime behaviour

- Scalar sensors run every `SENSOR_INTERVAL_SECONDS`.
- The camera captures at most once every five minutes. That interval is fixed
  in the timelapse component, not exposed in menuconfig.
- A sensor that fails to initialise is skipped; the remaining sensors keep
  publishing. Look for `skipping sensor` in the log.
- The WebSocket client reconnects automatically after a disconnect.
- The camera sends one image per WebSocket message. After saving an image, the
  server stops reading that connection but does not explicitly close it. Do not
  depend on a close frame or acknowledgement; a camera client should open a
  fresh connection for each image and close it when done.

## Common configuration errors

| Symptom | Likely cause |
| --- | --- |
| `WEBSOCKET_URL` empty or connects to port 80 | Wrong URL; include `ws://`, port, and `/api/websocket/sensors`. |
| WiFi connects but no readings | Server unreachable, firewall blocking port 8249, or wrong ID. |
| Sensor never starts | GPIO or I2C pin left at `0`, or the SHT/BH target ID left at `0`. A DS18B20 with system ID `0` still runs but writes orphan readings. |
| I2C sensors missing | `I2C_DATA_GPIO` or `I2C_CLOCK_GPIO` is `0`, or the bus is miswired. |
| Camera capture fails | No PSRAM, wrong board pin mapping, or insufficient power. |
| Flash fails | Missing bootloader mode, bad cable, or serial permissions. |

See [Troubleshooting](../troubleshooting) and
[Sensor Data Protocol](../sensor-data-protocol).
