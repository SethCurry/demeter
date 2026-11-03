---
title: "Getting Started"
summary: "Understand what Demeter does today, how the pieces fit together, and which guide to read next."
description: "Understand what Demeter monitors, how sensors, server, web panel, and Home Assistant fit together, and which guide to read next."
date: 2023-09-07T16:13:18+02:00
lastmod: 2026-11-03T00:00:00+00:00
draft: false
weight: 1
toc: true
---

## What is Demeter?

Demeter is a self-hosted, open-source monitoring and record-keeping system for
hydroponics. It brings three things together:

- **Monitoring** &mdash; continuous readings of enclosure air, reservoir water,
  and plant-site light, with history you can graph and review.
- **Timelapse photography** &mdash; camera frames captured by an ESP32 and
  played back as a timelapse.
- **A structured grow record** &mdash; enclosures, systems, flows, plant sites,
  plants, taxonomy, and timestamped notes.

Demeter currently focuses on **sensing and recording**. It does not yet switch
pumps, top up reservoirs, or schedule drains. The repository's roadmap mentions
maintenance automation, and some early internal packages exist, but there is
no user-facing automation feature in the current server or web panel.

> **Status:** Demeter is under heavy construction and not yet ready to deploy.
> An initial usable release is the immediate target &mdash; expect rough edges
> and breaking changes until then. In particular, the API has no
> authentication, several metrics are stored but not exposed, and the web panel
> is view/create-only for most records.

## How it fits together

Demeter is split into three cooperating pieces:

1. **Sensors (ESP32 firmware)** &mdash; small ESP32 devices with probes and
   cameras. They sample conditions and stream readings and images to the
   server over a WebSocket.
2. **The Demeter server** &mdash; a Go service that receives sensor data, stores
   it in an embedded SQLite database, serves the REST API, and hosts the web
   panel. A Home Assistant custom integration can read the same API.
3. **The web control panel** &mdash; a React + TypeScript single-page app that
   shows the dashboard, list and detail pages for enclosures, systems, flows,
   plants, genera, and species, and a Photos page. The Plants list is organized
   around plant sites.

```mermaid
architecture-beta
    group esp(server)[ESP32]
    group demeter(server)[Demeter]
    service user(cloud)[User]

    service api(server)[REST API] in demeter
    service db(disk)[SQLite] in demeter
    api:R --> L:db

    service homeassistant(server)[Home Assistant]
    api:L --> R:homeassistant

    junction websocket in esp

    service camera(disk)[Camera] in esp
    camera:T -- B:websocket

    service thermometer(disk)[Thermometer] in esp
    thermometer:T -- B:websocket

    service hygrometer(disk)[Hygrometer] in esp
    hygrometer:T -- B:websocket

    user:B --> T:api
    user:B --> T:homeassistant

    websocket:T --> B:api

    align row homeassistant api db
    align row thermometer camera hygrometer
    align column user api websocket
```

The server also serves the built web panel at `/`, so a browser talks to the
same process and the same port as the API and sensors.

## What it measures

| Domain | Readings | In the UI today? |
| --- | --- | --- |
| Water chemistry | pH, EC, dissolved oxygen | pH and EC yes; oxygen stored only |
| Water physical | temperature, water proximity/depth, flow rate | temperature yes; proximity and flow rate stored only |
| Environment | air temperature, air humidity, light intensity | all three yes |

The sensor WebSocket accepts all of these. REST read endpoints and charts
currently exist for air temperature, air humidity, pH, EC, water temperature,
and plant-site lux. Oxygen, proximity, and flow-rate readings are persisted in
SQLite but have no read endpoint or UI yet.

Demeter does **not** currently derive reservoir volume from depth. The
`system_water_proximity` table stores a depth value, but there is no reservoir
dimension model and no volume calculation.

## The domain model

Demeter organizes a grow as a hierarchy. Each level owns the measurements that
make sense at its scope:

```text
Enclosure                 air temperature, air humidity, notes
└── System                pH, EC, oxygen, water temperature, proximity, notes
    └── Flow              flow rate, notes; may nest under another flow
        └── Plant Site    lux, x/y/z coordinates
            └── Plant     planted date, species, notes
```

- **Enclosure** &mdash; the physical room, tent, or cabinet. Enclosures share
  air conditions, so air temperature and humidity live here.
- **System** &mdash; a reservoir and everything that shares its water. Water
  chemistry and water temperature live here.
- **Flow** &mdash; one water loop inside a system. Flows can nest under a parent
  flow and carry flow-rate readings.
- **Plant site** &mdash; one position on a flow, with integer X/Y/Z coordinates
  and a light reading.
- **Plant** &mdash; an individual specimen assigned to a site, optionally
  classified by species and genus.

The sidebar groups these entities under **Hydroponics** and **Taxonomy**
rather than drawing the full ownership tree; plant sites appear in the
**Plants** table, which is organized by site.

## The web panel at a glance

- **Dashboard** &mdash; counts of enclosures, systems, flows, and plant sites,
  plus a systems-status list.
- **Enclosures** &mdash; air-temperature and humidity charts, systems, flows,
  plants, and notes.
- **Systems** &mdash; pH, EC, and water-temperature charts, flows, plant sites,
  and notes.
- **Flows** &mdash; nested loops, plant sites, a 3D layout, plants, and notes.
- **Plants** &mdash; a site-centric table with coordinates, occupant, species,
  and planting date.
- **Genera / Species** &mdash; simple taxonomy management.
- **Photos** &mdash; playable MJPEG timelapses from cameras that have sent
  frames.
- **Notes** &mdash; add and read notes on enclosure, system, and flow pages.

The panel has no edit or delete controls yet; use the REST API for those
operations. Charts show history, not a live-updating stream. See
[Using the Web Control Panel](../using-the-web-control-panel).

## A minimal path to a working install

1. **Build and run the server** &mdash; see
   [Installing the Server](../installing-the-server).
2. **Create your topology** &mdash; an enclosure, a system, a flow, and a plant
   site. See [Enclosures](../enclosures),
   [Systems and Water Chemistry](../systems-and-water-chemistry), and
   [Flows and Plant Sites](../flows-and-plant-sites).
3. **Check the pipeline without hardware** &mdash; run the built-in simulator
   against target ID `1`.
4. **Build a sensor node** &mdash; wire the probes and flash the firmware. See
   [ESP32 Hardware](../esp32-hardware) and
   [Building and Flashing the ESP32 Firmware](../building-and-flashing-the-esp32-firmware).
5. **Confirm readings appear** &mdash; use the raw and metric endpoints or the
   detail pages.
6. **Optional: add Home Assistant** &mdash; see
   [Connecting to Home Assistant](../connecting-to-home-assistant).

## Guide map

| Guide | What it covers |
| --- | --- |
| [Installing the Server](../installing-the-server) | Build, run, Docker, first checks. |
| [Connecting to Home Assistant](../connecting-to-home-assistant) | Custom integration and entities. |
| [Using the Web Control Panel](../using-the-web-control-panel) | Pages, charts, and limitations. |
| [Enclosures](../enclosures) | Air conditions and enclosure ownership. |
| [Systems and Water Chemistry](../systems-and-water-chemistry) | Reservoirs, pH, EC, water temperature. |
| [Flows and Plant Sites](../flows-and-plant-sites) | Nested loops, coordinates, 3D layout. |
| [Plants and Taxonomy](../plants-and-taxonomy) | Plants, genera, species, lux. |
| [Notes](../notes) | Note types and current scoping caveats. |
| [Photos and Timelapse](../photos-and-timelapse) | Camera frames, storage, MJPEG playback. |
| [ESP32 Hardware](../esp32-hardware) | Boards, wiring, sensors. |
| [Building and Flashing the ESP32 Firmware](../building-and-flashing-the-esp32-firmware) | ESP-IDF, menuconfig, flash. |
| [Sensor Data Protocol](../sensor-data-protocol) | Exact WebSocket wire format. |
| [REST API](../rest-api) | Endpoints, shapes, metric queries. |
| [Troubleshooting](../troubleshooting) | Diagnose common failures. |
| [Operating and Maintaining Demeter](../operating-and-maintaining-demeter) | Services, backups, upgrades, docs. |

## Before you expose it to a network

Demeter has **no authentication**. Anyone who can reach port 8249 can read
data, create and delete records, inject sensor readings, and download
timelapse frames. Keep it on a trusted network, bind it to a private address, or
put an authenticating reverse proxy in front of it. See
[Operating and Maintaining Demeter](../operating-and-maintaining-demeter#exposing-the-port-safely).
