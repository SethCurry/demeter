---
title: "Systems and Water Chemistry"
summary: "Model a reservoir, create systems inside an enclosure, and work with pH, EC, and water-temperature readings."
description: "Create reservoirs, assign them to enclosures, interpret pH, EC, and water-temperature readings, and understand current unit conventions."
date: 2023-09-07T16:13:18+02:00
lastmod: 2026-11-03T00:00:00+00:00
draft: false
weight: 6
toc: true
---

A **system** is a hydroponic reservoir and everything that shares its water.
Systems are where Demeter stores water chemistry and water physical
measurements.

<figure>
  <img src="/images/sample.jpg" alt="System list and system detail page" width="1200" height="675" loading="lazy">
  <figcaption>The Systems page lists reservoirs and their enclosure. Replace this placeholder with a screenshot.</figcaption>
</figure>

## What a system owns

A system belongs to exactly one **enclosure** and contains:

- **Flows** — the water loops fed by the reservoir.
- **Plant sites** — indirectly, through flows.
- **Plants** — indirectly, through plant sites.
- **pH** readings.
- **Electrical conductivity (EC)** readings.
- **Water temperature** readings.
- **Dissolved oxygen** readings (stored, but not yet exposed by the API).
- **Water proximity / depth** readings (stored, but not yet exposed by the
  API).
- **System notes**.

The hierarchy is:

```text
Enclosure
└── System (reservoir)
    └── Flow (water loop)
        └── Plant Site (x, y, z)
            └── Plant
```

## Create a system

1. Open **Hydroponics &rarr; Systems**.
2. Click **New System**.
3. Enter a unique **Name**, for example `Main Reservoir`.
4. Choose the **Enclosure** that contains it.
5. Click **Create**.

Choose an enclosure that already exists. Names must be unique across all
systems. The database rejects a duplicate, but the current system create/update
handlers return an empty `500` response rather than the intended `409`. The
current server does not enforce the enclosure foreign key, so the API can also
store a system that points at a nonexistent enclosure. Such a system will not
appear under any real enclosure. The UI picker only offers existing enclosures, so prefer it
when creating systems.

API example:

```sh
curl -X POST http://localhost:8249/api/systems \
  -H 'Content-Type: application/json' \
  -d '{"name":"Main Reservoir","enclosure_id":1}'
```

The response contains the new system `ID`, which sensors use as their target.

## The system list

The table shows each system's ID, name, enclosure, and number of flows. Names
link to the detail page. Use **Refresh** after changing the topology.

You can filter the list by enclosure through the API with
`GET /api/systems?enclosure_id=1`; the UI uses this on the enclosure detail
page.

## The system detail page

### Summary and related records

The description card shows the system ID and name, a link to its enclosure, and
counts of flows, plant sites, and plants. The page also lists the flows inside
the system, the plant sites attached to those flows, and the plants on those
sites.

### Latest readings

Three statistic cards show the most recent raw reading for:

- **pH**
- **EC**
- **Water Temperature** in degrees Celsius

A card shows **No readings yet** when no reading exists. The latest value is
simply the first row returned by the raw-reading endpoint, which is ordered
newest first and limited to 20 rows on this page.

### Charts

Each of the three cards also contains a small line chart below its latest
value:

- **pH**
- **EC**
- **Water Temperature**

These are raw-reading charts. They plot the same newest-first data as the
latest-value display: up to 20 readings per metric, with no time-range control
and no separate Refresh button. They do not use the interval-averaged metric
endpoints yet. For interval-averaged charts with selectable ranges, use the
enclosure air charts or query the metric endpoints directly. See
[Using the Web Control Panel](../using-the-web-control-panel#reading-the-charts)
and [REST API](../rest-api#metric-series).

### Notes

The **Notes** card lists system notes and provides **Add Note**. See
[Notes](../notes).

## Units and wire encoding

Demeter does not attach unit metadata to stored readings. The de facto
conventions in the firmware, UI, and Home Assistant integration are:

| Metric | Stored field | Displayed as | Wire encoding |
| --- | --- | --- | --- |
| pH | `ph` | pH | pH × 1000 |
| EC | `ec` | Web UI labels it **mS/cm**; Home Assistant labels it **µS/cm**; neither converts the number | EC × 1000 |
| Water temperature | `water_temperature_c` | °C | °C × 1000 |
| Dissolved oxygen | `oxygen` | no UI yet | value × 1000 |
| Water proximity | `proximity_m` | no UI yet | metres × 1000 |

No unit is stored with an EC reading. Decide which unit your probe reports and
keep it consistent across the sensor, the web panel, and Home Assistant. The
current web panel displays the bare number with an `mS/cm` suffix, while the
Home Assistant integration declares `µS/cm` without scaling it. Those two
labels disagree by a factor of 1000, so do not mix them.

For example, a pH of `6.42` is sent as `4:1:6420`: message type `4`, system
ID `1`, value `6420`. The server divides by `1000` before inserting the
reading.

The stock ESP32 firmware only measures water temperature with a DS18B20. pH,
EC, and dissolved-oxygen probes must be integrated by a custom sensor client
using the same WebSocket protocol. See
[Sensor Data Protocol](../sensor-data-protocol).

## Measurements without an API yet

The server accepts and persists dissolved oxygen and water proximity
readings, but the current REST API exposes only pH, EC, and water temperature
for systems. There is no chart or latest-value card for oxygen or proximity.
If you need those readings today, query the SQLite database directly or add an
API endpoint.

## Rename or delete a system

Use the API while the UI has no edit/delete controls:

```sh
# Rename
curl -X PUT http://localhost:8249/api/systems/1 \
  -H 'Content-Type: application/json' \
  -d '{"name":"Veg Reservoir","enclosure_id":1}'

# Delete
curl -X DELETE http://localhost:8249/api/systems/1
```

Deleting a system does not cascade to flows, plant sites, plants, or readings.
Because foreign keys are not enforced, the delete succeeds even when dependants
exist and leaves them pointing at a missing system. Remove or re-parent
dependants first and back up the database.

## API endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/systems` | List systems; optional `enclosure_id` filter. |
| `POST` | `/api/systems` | Create a system. |
| `GET` | `/api/systems/{id}` | Get one system. |
| `PUT` | `/api/systems/{id}` | Rename/re-parent a system. |
| `DELETE` | `/api/systems/{id}` | Delete a system. |
| `GET` | `/api/systems/{id}/ph` | Latest raw pH readings. |
| `GET` | `/api/systems/{id}/ec` | Latest raw EC readings. |
| `GET` | `/api/systems/{id}/water-temperature` | Latest raw water temperatures. |
| `GET` | `/api/systems/{id}/metrics/ph` | Averaged pH series. |
| `GET` | `/api/systems/{id}/metrics/ec` | Averaged EC series. |
| `GET` | `/api/systems/{id}/metrics/water-temperature` | Averaged water-temperature series. |

See [REST API](../rest-api) for details.
