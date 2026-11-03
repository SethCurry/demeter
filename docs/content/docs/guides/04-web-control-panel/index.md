---
title: "Using the Web Control Panel"
summary: "Tour the React control panel: navigation, the dashboard, detail pages, charts, refresh behaviour, and the parts of the UI that are view-only."
date: 2023-09-07T16:13:18+02:00
lastmod: 2026-11-03T00:00:00+00:00
draft: false
weight: 4
toc: true
---

The Demeter server embeds the compiled web control panel, so there is nothing
extra to install. Once the server is running, open its address in a browser:

```text
http://localhost:8249
```

The panel is a single-page React application. It reads and writes the same
REST API documented in [REST API](../rest-api), and it is served by the Go
binary from `/`.

<figure>
  <img src="/images/sample.jpg" alt="Demeter control panel" width="1200" height="675" loading="lazy">
  <figcaption>The dashboard is the landing page. Replace this placeholder with a screenshot of your installation.</figcaption>
</figure>

> **No authentication:** the current build does not have a login screen or API
> tokens. Anyone who can reach port `8249` can read and modify the plant data.
> Keep Demeter on a trusted network, or put it behind a reverse proxy that
> provides authentication and TLS.

## Layout and navigation

The left sidebar contains the whole application. On narrow screens the
sidebar collapses behind the menu button.

| Menu item | Page | What it is for |
| --- | --- | --- |
| **Dashboard** | `/dashboard` | Counts of enclosures, systems, flows, and plant sites, plus a systems list. |
| **Hydroponics &rarr; Enclosures** | `/enclosures` | Physical grow spaces and their air conditions. |
| **Hydroponics &rarr; Systems** | `/systems` | Reservoirs and their water chemistry. |
| **Hydroponics &rarr; Flows** | `/flows` | Water loops within a system, including nested loops. |
| **Hydroponics &rarr; Plants** | `/plants` | Plant sites, their coordinates, and the plants assigned to them. |
| **Taxonomy &rarr; Genera** | `/plant-genera` | Plant genera. |
| **Taxonomy &rarr; Species** | `/plant-species` | Plant species, optionally grouped under a genus. |
| **Photos** | `/photos` | Timelapse recordings received from cameras. |

The paths in the table are client-side routes. The server only serves the app
at `/`; opening a deep link such as `http://localhost:8249/enclosures`
directly returns a plain `404 page not found`. Always start at the server root
and navigate with the sidebar, or bookmark the root rather than an inner page.
A hard browser reload on an inner page has the same problem.

The entity list and detail pages have a **Refresh** button. The Dashboard and
Photos pages do not; reload the browser or revisit those pages. The panel does
not maintain a live push connection for ordinary pages: it fetches when the
page opens and again when you click **Refresh**. Reload the browser page if the
application itself has been updated by a server upgrade.

## The dashboard

The dashboard is intentionally an overview rather than a control surface. It
shows:

- **Enclosures**, **Systems**, **Flows**, and **Plant Sites** totals.
- A **Systems Status** list. Each entry links to the system detail page and
  shows the system's enclosure and how many flows belong to it.

It does not currently show live pH, EC, temperature, or light values. Open an
enclosure, system, flow, or plant detail page for readings and charts.

## List pages and detail pages

Most entity list pages follow the same pattern:

1. A heading and short description of the concept.
2. **Refresh** and a primary **New …** button.
3. Summary cards and/or a table.
4. Names and IDs link to the matching detail page.

Detail pages add:

- A breadcrumb-style back button.
- A **Refresh** button that reloads every dataset used by the page.
- Description cards with IDs, names, and parent/child relationships.
- Charts and note tables where the concept supports them.

The UI currently creates records but does not expose edit or delete buttons.
The API does support `PUT` and `DELETE` for most resources, so use the
[REST API](../rest-api) when you need to rename, move, or remove a record.

## Reading the charts

Charts differ by detail page.

**Enclosure detail pages** have two metric charts: air temperature and air
humidity. Each metric chart has:

- A time-range control with **1h**, **6h**, **24h**, **7d**, **30d**, and
  **Custom** ranges. The default is the last 24 hours.
- A **Refresh** button beside the range control.
- A line chart whose points are interval averages, not every raw reading.

The browser requests a number of buckets that keeps the chart readable
(roughly 120 points) and never uses a bucket shorter than one minute. The
server then splits the selected window into that many equal intervals and
returns the average of all readings in each non-empty interval. Gaps in the
line therefore mean "no readings in that interval", not zero.

**System detail pages** show pH, EC, and water-temperature cards with a small
raw-reading line chart inside each card. Each chart uses the same newest-first
data as its latest-value display: up to 20 raw readings, with no range control
and no separate Refresh button. These are raw charts, not interval averages.

**Plant detail pages** use a separate **Lux** line chart. It plots the raw lux
readings returned for the plant site (newest first, up to 100) and has no
time-range control. The chart reverses its time labels to oldest-first but does
not reverse the values, so the labels and points may not line up; treat it as a
rough history view until that is fixed. Flow detail pages do not currently show
a flow-rate chart.

For the exact query parameters and response shape, see
[REST API](../rest-api#metric-series).

## Common tasks

### Create your first topology

A useful order is:

1. Create an **Enclosure**.
2. Create a **System** inside that enclosure.
3. Create a **Flow** inside that system.
4. Create **Plant Sites** at coordinates within the flow.
5. Optionally create **Genera** and **Species**, then add **Plants** to sites.

The detail pages link to their parents and children, so you can start from any
level and navigate sideways. See [Enclosures](../enclosures),
[Systems](../systems-and-water-chemistry), [Flows and Plant Sites](../flows-and-plant-sites),
and [Plants and Taxonomy](../plants-and-taxonomy).

### Check why a chart is empty

An empty chart usually means one of:

- No sensor has reported that metric for the selected entity ID.
- On an enclosure chart, the selected time window contains no readings.
- On a system or plant chart, the page's raw-reading query returned no rows.
- The entity does not exist, or the sensor is configured with the wrong
  `target_id`.

System charts and latest-value cards come from the same raw readings, so if a
card says **No readings yet**, its chart is empty too. Plant pages show a raw
**Lux** chart. Enclosure pages have metric charts only and no separate
latest-value card. See [Troubleshooting](../troubleshooting).

## What is not in the panel

This is a deliberately honest list for the current build:

- No user accounts or permissions.
- No edit/delete controls on list and detail pages.
- No live WebSocket updates; use **Refresh**.
- No page for changing server-side settings. The `Settings.tsx` source file
  exists in the repository but is not routed or shown in the menu, so it is
  not a supported user-facing feature yet.
- No working capture button. The Photos page shows a **New Timelapse** button,
  but it has no handler. Cameras capture on their own schedule; the Photos
  page lists what has already arrived.
- No charts for dissolved oxygen, water proximity, or flow rate yet, even
  though the server accepts and stores those readings.
