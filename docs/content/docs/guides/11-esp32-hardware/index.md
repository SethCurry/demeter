---
title: "ESP32 Hardware"
summary: "Choose a board, wire the DS18B20, SHT4x, BH1750, and camera, and understand the GPIO conventions used by the stock firmware."
date: 2023-09-07T16:13:18+02:00
lastmod: 2026-11-03T00:00:00+00:00
draft: false
weight: 11
toc: true
---

Demeter's reference sensor node is an **ESP32-S3** development board with a
camera and up to three environmental sensors. This page covers the hardware
side. For building and flashing the firmware, see
[Building and Flashing the ESP32 Firmware](../building-and-flashing-the-esp32-firmware).

<figure>
  <img src="/images/sample.jpg" alt="ESP32 sensor node wiring" width="1200" height="675" loading="lazy">
  <figcaption>A typical ESP32-S3 camera board with a DS18B20, SHT4x, and BH1750. Replace this placeholder with a wiring photo or diagram.</figcaption>
</figure>

## Recommended board

Use an ESP32-S3 board with:

- **PSRAM** — at least 2 MB, ideally 8 MB. Camera frames do not fit in internal
  RAM.
- **A camera connector** compatible with the `esp32-camera` component for your
  board. `sdkconfig.default` requests PSRAM support and PSRAM DMA, but those
  defaults only apply to a freshly generated configuration; check menuconfig
  after upgrading or reusing an existing `sdkconfig`.
- **3.3 V logic** and enough free GPIO pins for the sensors you plan to use.
- **USB** for flashing and serial monitoring.

The firmware is not tied to one board, but the camera pins in this repository
are hard-coded for one ESP32-S3 camera board. They are **not** selected by
menuconfig. The defaults live in
`esp-sensor/components/timelapse-sensor/include/timelapse-sensor.h`:

| Signal | GPIO |
| --- | --- |
| PWDN / RESET | not connected (`-1`) |
| XCLK | 15 |
| SIOD (SDA) | 4 |
| SIOC (SCL) | 5 |
| VSYNC | 6 |
| HREF | 7 |
| PCLK | 13 |
| D0–D7 | 11, 9, 8, 10, 12, 18, 17, 16 |

If your camera board uses different pins, edit those defines and rebuild. Do
not connect a camera using a different pinout without changing the firmware;
the image will be corrupt or the camera will fail to initialise.

An ESP32 without PSRAM can still run the scalar sensors, but not the camera
path reliably.

## Sensors

### DS18B20 — water temperature

A waterproof 1-Wire probe for a reservoir. It reports water temperature for
one **system**.

- **Bus:** 1-Wire, one data GPIO.
- **Supply:** 3.3 V (or 5 V with a level shifter, depending on the board and
  probe).
- **Pull-up:** a 4.7 kΩ resistor from the data line to 3.3 V is required.
  Many waterproof probes do not include one.
- **Addressing:** the firmware uses the first device found on the bus. Use one
  DS18B20 per ESP32 unless you modify the firmware for multiple addresses.
- **Cable length:** long runs pick up noise. Keep the pull-up near the board
  and avoid running the probe cable next to mains or pump wiring.

Set `DS18B20_GPIO` to the data pin and `DS18B20_SYSTEM_ID` to the reservoir's
numeric ID. Setting `DS18B20_GPIO` to `0` disables the sensor.

### SHT4x — air temperature and humidity

A small I2C sensor for the enclosure's air. It reports two readings for one
**enclosure**: temperature and relative humidity.

- **Bus:** I2C, shared with the BH1750.
- **Supply:** 3.3 V.
- **Address:** fixed by the SHT4x variant; no address configuration is needed.
- **Placement:** away from the reservoir, lights, and the ESP32's own heat.
  Radiated heat will bias the reading.

Set `SHT_ENCLOSURE_ID` to the enclosure's numeric ID. Setting it to `0`
disables the sensor. The I2C bus is enabled when `I2C_DATA_GPIO` is non-zero;
both `I2C_CLOCK_GPIO` and `I2C_DATA_GPIO` must be set to valid pins.

### BH1750 — plant-site light

An I2C ambient-light sensor for one **plant site**. It reports illuminance in
lux.

- **Bus:** I2C, shared with the SHT4x.
- **Supply:** 3.3 V; most breakout boards include the required pull-ups.
- **Placement:** at canopy height, facing the light, without shadowing the
  plant.
- **Address:** the common breakout defaults to `0x23`. The firmware uses the
  default address.

Set `BH_PLANT_SITE_ID` to the plant site's numeric ID. Setting it to `0`
disables the sensor.

### Camera

The camera captures timelapse frames. It does not need a target ID; frames are
stored under the device's MAC address.

- Use a camera module supported by `esp32-camera` for your board.
- Check the hard-coded pin table above before connecting a module.
- The camera needs PSRAM and a stable 3.3 V supply. Brownouts during capture
  usually mean the USB supply or regulator is inadequate.
- Focus is fixed on most modules. The firmware attempts autofocus when the
  sensor supports it.
- There is no menuconfig switch for the camera. The firmware always registers
  the timelapse sensor; if camera initialisation fails, that sensor is skipped
  and the rest keep running.

## I2C bus

The SHT4x and BH1750 share one bus:

| Signal | Configuration |
| --- | --- |
| SCL | `I2C_CLOCK_GPIO` |
| SDA | `I2C_DATA_GPIO` |

Both sensors are 3.3 V devices. If you use breakout boards, check whether
they already include pull-ups; adding a second set in parallel can make the
bus too strong. A single pair of 4.7 kΩ pull-ups to 3.3 V is typical.

The firmware initialises the I2C driver only when `I2C_DATA_GPIO` is non-zero.
If you enable the SHT4x or BH1750 but leave the data pin at `0`, the sensor
will not start.

## GPIO 0 disables a sensor

The firmware uses `0` as its "not configured" sentinel:

| Setting | Disabled when |
| --- | --- |
| `DS18B20_GPIO` | `0` |
| `I2C_DATA_GPIO` | `0` (disables the whole I2C bus) |
| `SHT_ENCLOSURE_ID` | `0` |
| `BH_PLANT_SITE_ID` | `0` |

Do not use GPIO 0 for a real sensor unless you also change the firmware's
compile-time checks. On many boards GPIO 0 is a boot-strapping pin and is a
poor choice anyway.

## Power and enclosure

- Power the board from a stable 5 V USB supply, not from a shared pump relay.
- Keep the ESP32 outside splash range. A ventilated project box with cable
  glands works well.
- Route the DS18B20 cable away from AC mains and pump power cables.
- Do not seal the SHT4x in a waterproof box; it must sample the enclosure air.
- If the board and sensors are far apart, keep I2C short (under about a metre)
  and use shielded cable for the 1-Wire probe.

## A minimal build

For a first installation you only need:

- ESP32-S3 camera board.
- DS18B20 with a 4.7 kΩ pull-up, in the reservoir.
- SHT4x inside the enclosure.
- 3.3 V, ground, SCL, and SDA between the board and the I2C sensors.

Add the BH1750 once you have a plant site and want light history. The camera is
optional; without it, everything except [Photos and Timelapse](../photos-and-timelapse)
works normally.

## Where to go next

- [Building and Flashing the ESP32 Firmware](../building-and-flashing-the-esp32-firmware) — configure
  the pins and target IDs, then build.
- [Sensor Data Protocol](../sensor-data-protocol) — understand exactly what
  the device sends.
- [Enclosures](../enclosures), [Systems and Water Chemistry](../systems-and-water-chemistry),
  and [Plants and Taxonomy](../plants-and-taxonomy) — create the target IDs the
  firmware needs.
