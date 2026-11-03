---
title: "Enclosures"
summary: "Create and manage grow spaces, understand what an enclosure owns, and read its air-temperature and humidity history."
date: 2023-09-07T16:13:18+02:00
lastmod: 2026-11-03T00:00:00+00:00
draft: false
weight: 5
toc: true
---

An **enclosure** is the physical room, tent, cabinet, or greenhouse that holds
your hydroponic equipment. It is the top level of the Demeter hierarchy and
the scope for air conditions that are shared by everything inside it.

<figure>
  <img src="/images/sample.jpg" alt="Enclosure list and detail page" width="1200" height="675" loading="lazy">
  <figcaption>The Enclosures page shows one card per enclosure and a table with system counts. Replace this placeholder with a screenshot.</figcaption>
</figure>

## What an enclosure owns

An enclosure contains:

- **Systems** — one or more reservoirs inside the enclosure.
- **Flows** — indirectly, through the systems.
- **Plant sites** — indirectly, through the flows.
- **Plants** — indirectly, through the plant sites.
- **Air temperature** readings from a sensor such as an SHT4x.
- **Air humidity** readings from the same sensor.
- **Enclosure notes**.

Air temperature and humidity belong to the enclosure rather than to a single
system because they describe the shared environment. Water chemistry belongs
to a system instead — see [Systems](../systems-and-water-chemistry).

## Create an enclosure

1. Open **Hydroponics &rarr; Enclosures**.
2. Click **New Enclosure**.
3. Enter a unique **Name**, for example `Vegetable Tent`.
4. Click **Create**.

Names must be unique across all enclosures. The database rejects a duplicate,
but the current enclosure create/update handlers return an empty `500` response
rather than the intended `409`. If the request fails and the name is already in
use, choose a different name.

You can also create an enclosure with the API:

```sh
curl -X POST http://localhost:8249/api/enclosures \
  -H 'Content-Type: application/json' \
  -d '{"name":"Vegetable Tent"}'
```

The response includes the new `ID`. Write it down: sensors are configured with
numeric target IDs, not names.

## The enclosure list

Each enclosure appears as a card and in the table below it. Both open the
detail page when clicked. The list shows:

- The enclosure **ID** and **Name**.
- The number of **Systems** inside it.

Use **Refresh** after creating systems in another tab or after a sensor starts
reporting.

## The enclosure detail page

The detail page is organised into four areas.

### Summary

The description card shows the enclosure's ID and name, plus counts of its
systems, flows, plant sites, and plants. The counts are derived by walking the
hierarchy, so a flow appears under the enclosure of the system it belongs to.

### Air conditions

Two chart cards show the enclosure's air history:

- **Air Temperature** in degrees Celsius.
- **Air Humidity** as relative humidity percent.

The page does not show a separate latest-value card for either metric. Hover a
chart point to see the interval value and time. The charts use the
interval-averaged metric endpoints. Pick a range from **1h** to **30d**, or
choose **Custom** to set exact start and end times. A gap means no reading was
recorded in that interval.

### Systems, flows, and plants

The page lists the systems inside the enclosure and the flows that belong to
those systems. It also lists the plants found by following
system &rarr; flow &rarr; plant site &rarr; plant. Links take you to the
matching detail page.

The plant count and plant card are taxonomy-limited. They use
`GET /api/enclosures/{id}/plants`, whose query inner-joins `plant_species` and
`plant_genus`. A plant with no species, or a species with no genus, is omitted
from this card and from the count. The **Plants** list and the plant detail
page still show unclassified plants.

### Notes

The **Notes** card lists enclosure notes and has an **Add Note** button. See
[Notes](../notes) for behaviour and current limitations.

## Connecting an air sensor

The stock ESP32 firmware supports an SHT4x over I2C. In the firmware
configuration, set:

- `SHT_ENCLOSURE_ID` to the numeric ID of this enclosure.
- `I2C_CLOCK_GPIO` and `I2C_DATA_GPIO` to the pins used by the sensor.

The device sends two measurements per sample: temperature as message type
`2` and humidity as message type `3`. The server divides the received integer
values by `1000` before storing them. See
[Connecting ESP32 Sensors](../esp32-hardware) and
[Sensor Data Protocol](../sensor-data-protocol).

## Rename or delete an enclosure

The current web panel does not expose edit or delete buttons. Use the API:

```sh
# Rename
curl -X PUT http://localhost:8249/api/enclosures/1 \
  -H 'Content-Type: application/json' \
  -d '{"name":"New Name"}'

# Delete
curl -X DELETE http://localhost:8249/api/enclosures/1
```

Deleting an enclosure does not clean up systems, flows, plant sites, or
readings. Because foreign keys are not enforced, the delete succeeds even when
dependants exist and leaves dangling IDs behind. Remove or re-parent dependent
records first, and take a database backup before bulk changes. See
[Operating and Maintaining Demeter](../operating-and-maintaining-demeter).

## API endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/enclosures` | List enclosures. |
| `POST` | `/api/enclosures` | Create an enclosure. |
| `GET` | `/api/enclosures/{id}` | Get one enclosure. |
| `PUT` | `/api/enclosures/{id}` | Rename an enclosure. |
| `DELETE` | `/api/enclosures/{id}` | Delete an enclosure. |
| `GET` | `/api/enclosures/{id}/plants` | List plants in the enclosure. |
| `GET` | `/api/enclosures/{id}/air-temperature` | Latest raw temperature readings. |
| `GET` | `/api/enclosures/{id}/air-humidity` | Latest raw humidity readings. |
| `GET` | `/api/enclosures/{id}/metrics/air-temperature` | Averaged temperature series. |
| `GET` | `/api/enclosures/{id}/metrics/air-humidity` | Averaged humidity series. |

See [REST API](../rest-api) for query parameters, response shapes, and
examples.
