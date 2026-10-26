---
title: "Getting Started"
summary: ""
date: 2023-09-07T16:13:18+02:00
lastmod: 2023-09-07T16:13:18+02:00
draft: false
weight: 1
toc: true
---


## What is Demeter?

Demeter is a self-hosted, open-source control and monitoring system for
hydroponics. It brings together the three things a grower needs in one place:

- **Monitoring** &mdash; continuous readings of the conditions that matter to
  plants, with a history you can graph and review.
- **Timelapse photography** &mdash; automated capture of the whole system and of
  roots, so you can watch growth and catch problems early.
  drains without manual intervention.

The goal is to make running a hydroponic system less of a chore and more of a
watched, data-driven process: instead of checking reservoirs by hand and
wondering whether something drifted overnight, Demeter records everything and
lets you act on it from a single dashboard.

> **Status:** Demeter is under heavy construction and not yet ready to deploy.
  An initial usable release is the immediate target &mdash; expect rough edges
  and breaking changes until then.

## How it fits together

Demeter is split into three cooperating pieces:

1. **Sensors (ESP32 firmware)** &mdash; small ESP32 devices with attached
   probes and cameras. They sample conditions in the grow space and stream
   readings and images back over a WebSocket connection to the server.
2. **The Demeter server** &mdash; a Go service that receives
   sensor data, stores it in an embedded SQLite database, exposes a REST API
   for querying and control, and integrates with [Home Assistant](https://www.home-assistant.io/) for people who already run a smart home.
3. **The web control panel** &mdash; a React + TypeScript single-page app
   (Ant Design) that drives the API and gives you a live dashboard plus pages
   for every part of your setup.


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


## What it measures

A Demeter installation tracks the full set of variables that affect a
hydroponic reservoir and its surroundings:

| Domain            | Readings                                            |
|-------------------|-----------------------------------------------------|
| Water chemistry   | pH, electrical conductivity (EC), dissolved O₂  |
| Water physical    | temperature, level, flow rate                       |
| Environment       | air temperature, air humidity, light intensity      |

From water depth and the reservoir's dimensions, Demeter also derives the
total volume of the reservoir, so you can track consumption and top-ups in
real units rather than abstract levels.

## The domain model

Demeter organizes a grow around four nested concepts. Each one owns the
measurements that make sense at its scope:

- **Enclosure** the physical space everything sits in. Enclosures share
  air qualities: air temperature and humidity.
- **System** defined by a shared reservoir. Systems own the water
  chemistry and physical readings: pH, EC, O₂, water temperature, and level.
- **Flow** a loop sharing a single flow of water. Flows carry the flow
  rate and can be nested under a parent flow, with a count of plant sites.
- **Plant site** a single location within a loop. Each site has a
  position, lux, species, and planted date.

This hierarchy is mirrored directly in the web UI's sidebar, so navigating the
control panel matches how you think about the physical system.

## What it can do for you

- See a live dashboard of every system with average pH, EC, and light, plus
  per-system status.
- Browse history and graph metrics over time for any enclosure, system, flow,
  or plant site.
- Capture and review timelapse photos of whole systems and of roots.
- Automate reservoir top-ups and schedule routine drains.
- Pull everything into Home Assistant alongside the rest of your home.

The rest of this guide walks through installing the server, flashing a sensor,
and pointing the web panel at your backend.
