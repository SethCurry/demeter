# Demeter

Demeter is a self-hosted, open-source control and monitoring system for
hydroponics. It brings together the things a grower needs in one place:

- **Monitoring** &mdash; continuous readings of the conditions that matter to
  plants, with a history you can graph and review.
- **Timelapse photography** &mdash; automated capture of the whole system and
  of roots, so you can watch growth and catch problems early.
- **Automated maintenance** &mdash; top up reservoirs and schedule routine
  drains without manual intervention.

The goal is to make running a hydroponic system less of a chore and more of a
watched, data-driven process: instead of checking reservoirs by hand and
wondering whether something drifted overnight, Demeter records everything and
lets you act on it from a single dashboard.

> **Status:** Demeter is under heavy construction and not yet ready to deploy.
> An initial usable release is the immediate target &mdash; expect rough edges
> and breaking changes until then.

## How it fits together

Demeter is split into three cooperating pieces:

1. **Sensors (ESP32 firmware)** &mdash; small ESP32 devices with attached
   probes and cameras. They sample conditions in the grow space and stream
   readings and images back over a WebSocket connection to the server.
2. **The Demeter server** &mdash; a Go service that receives sensor data,
   stores it in an embedded SQLite database, exposes a REST API for querying
   and control, and integrates with
   [Home Assistant](https://www.home-assistant.io/) for people who already run
   a smart home.
3. **The web control panel** &mdash; a React + TypeScript single-page app
   (Ant Design) that drives the API and gives you a live dashboard plus pages
   for every part of your setup.

```
ESP32 sensors ──websocket──▶ Demeter server ──REST──▶ Web panel
                                  │
                                  ├── SQLite (embedded)
                                  └── Home Assistant
```

## What it measures

A Demeter installation tracks the full set of variables that affect a
hydroponic reservoir and its surroundings:

| Domain          | Readings                                         |
|-----------------|--------------------------------------------------|
| Water chemistry | pH, electrical conductivity (EC), dissolved O₂   |
| Water physical  | temperature, level, flow rate                    |
| Environment     | air temperature, air humidity, light intensity   |

From water depth and the reservoir's dimensions, Demeter also derives the
total volume of the reservoir, so you can track consumption and top-ups in
real units rather than abstract levels.

## The domain model

Demeter organizes a grow around four nested concepts. Each one owns the
measurements that make sense at its scope:

- **Enclosure** &mdash; the physical space everything sits in. Enclosures
  share air qualities: air temperature and humidity.
- **System** &mdash; defined by a shared reservoir. Systems own the water
  chemistry and physical readings: pH, EC, O₂, water temperature, and level.
- **Flow** &mdash; a loop sharing a single flow of water. Flows carry the
  flow rate and can be nested under a parent flow, with a count of plant
  sites.
- **Plant site** &mdash; a single location within a loop. Each site has a
  position, lux, species, and planted date.

This hierarchy is mirrored directly in the web UI's sidebar, so navigating the
control panel matches how you think about the physical system.

## Repository layout

```
cmd/demeter      Go server entrypoint (HTTP + WebSocket, sensor simulator)
internal/
  embedded/      embedded SQLite schema and the built web UI
  httpd/         REST API and WebSocket handlers
  models/        sqlc-generated data access
  sensor/        WebSocket sensor connection/receiver
  dscript/       maintenance automation runtime
esp-sensor/      ESP32 firmware for sensor devices
web/             React + TypeScript control panel
python/          Home Assistant integration
docs/            Hugo-based documentation site
sql/             source SQL schema (copied into the build)
```

## Building and running

The repo uses [Task](https://taskfile.dev). Building the server embeds the
web UI and the SQL schema into the Go binary:

```sh
task build        # build web UI, copy assets, and build the demeter binary
./demeter server  # run the HTTP server on :8249
```

For development, the included taskfiles run the docs site, the web panel, and
an ESP firmware build separately:

```sh
task site   # docs site (Hugo)
task ui:dev # web control panel
task esp    # ESP32 firmware
```

The `demeter` binary also ships a synthetic sensor simulator for exercising
the backend without hardware:

```sh
./demeter simulate-sensor --addr localhost:8249
```

## Documentation

Full guides and sensor/controller reference docs live under
[`docs/`](docs/) and are published from the Hugo site in that directory. The
[Getting Started guide](docs/content/docs/guides/01-getting-started/index.md)
is the best place to begin.