---
title: "Guides"
summary: "Task-oriented guides for installing, using, and maintaining Demeter."
description: "Install Demeter, model your grow, connect ESP32 sensors, use the REST API, and keep a self-hosted hydroponics server running."
date: 2023-09-07T16:12:37+02:00
lastmod: 2026-11-03T00:00:00+00:00
draft: false
weight: 1
toc: true
---

These guides cover the whole of Demeter: the server, the web panel, the
domain model, the ESP32 firmware, the sensor wire protocol, the REST API, and
day-to-day operations.

If you are new to Demeter, start with
[Getting Started](getting-started). If you are looking for a specific
concept or task, use the table below.

## Start here

| Guide | What you will learn |
| --- | --- |
| [Getting Started](getting-started) | What Demeter does today, the architecture, and the domain model. |
| [Installing the Server](installing-the-server) | Build and run the server, or use the Docker dev stack. |
| [Connecting to Home Assistant](connecting-to-home-assistant) | Install the custom integration and use the sensor entities. |
| [Using the Web Control Panel](using-the-web-control-panel) | Navigate the UI, read charts, and understand its limits. |

## Model your grow

| Guide | What you will learn |
| --- | --- |
| [Enclosures](enclosures) | Grow spaces, air temperature and humidity, and sensor IDs. |
| [Systems and Water Chemistry](systems-and-water-chemistry) | Reservoirs, pH, EC, water temperature, and units. |
| [Flows and Plant Sites](flows-and-plant-sites) | Nested water loops, X/Y/Z layout, flow rate. |
| [Plants and Taxonomy](plants-and-taxonomy) | Plants, planting dates, genera, species, and lux. |
| [Notes](notes) | Observations, timestamps, and current note scoping. |
| [Photos and Timelapse](photos-and-timelapse) | Camera frames, storage, playback, and retention. |

## Hardware and integration

| Guide | What you will learn |
| --- | --- |
| [ESP32 Hardware](esp32-hardware) | Boards, wiring, GPIOs, and sensor placement. |
| [Building and Flashing the ESP32 Firmware](building-and-flashing-the-esp32-firmware) | ESP-IDF setup, menuconfig, build, flash, monitor. |
| [Sensor Data Protocol](sensor-data-protocol) | The exact WebSocket text and binary formats. |
| [REST API](rest-api) | Endpoints, response shapes, metric queries, and examples. |

## Operate and fix

| Guide | What you will learn |
| --- | --- |
| [Troubleshooting](troubleshooting) | Diagnose startup, networking, sensor, chart, and camera failures. |
| [Operating and Maintaining Demeter](operating-and-maintaining-demeter) | systemd, backups, upgrades, retention, security, docs builds. |

## A note on current limitations

Demeter is under active development. The guides describe the behaviour of the
code in this repository, including features that are incomplete:

- The API and sensor WebSocket have **no authentication**.
- The web panel creates records but generally does not edit or delete them.
- Dissolved oxygen, water proximity, and flow-rate readings are accepted and
  stored, but have no read endpoint or chart yet.
- Enclosure, system, and flow notes are not linked to their parent records.
- Timelapse storage is not on a persistent Docker volume by default.

When a guide says a feature is unavailable, treat that as the current
behaviour, not a permanent design decision.
