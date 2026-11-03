---
title: "REST API"
summary: "Use Demeter's HTTP API directly: conventions, authentication, response shapes, metric queries, and a complete endpoint index."
date: 2023-09-07T16:13:18+02:00
lastmod: 2026-11-03T00:00:00+00:00
draft: false
weight: 14
toc: true
---

The web control panel is a client of the same REST API that is available to
scripts, Home Assistant, and custom dashboards. This page documents the
conventions you need to use it correctly and indexes every endpoint.

<figure>
  <img src="/images/sample.jpg" alt="Interactive OpenAPI reference" width="1200" height="675" loading="lazy">
  <figcaption>The generated OpenAPI reference is available at /openapi.html. Replace this placeholder with a screenshot.</figcaption>
</figure>

An interactive reference generated from `docs/static/openapi.yaml` is served
at [`/openapi.html`](/openapi.html) when you run the documentation site. It is
useful for browsing request and response schemas; this page focuses on the
behaviour those schemas do not express.

## Base URL

The server listens on the address given to `demeter server` (default
`:8249`). With the defaults:

```text
http://localhost:8249
```

All API routes are under `/api`. The embedded web UI is served only at `/`, so
an unknown `/api` path returns a plain `404 page not found` rather than JSON or
HTML. Client-side routes such as `/enclosures` also return a plain `404` when
opened directly.

## Authentication

**There is no authentication.** Every API route, including writes, deletes,
the sensor WebSocket, and the timelapse stream, is open to anyone who can reach
the port. There are no users, tokens, cookies, or API keys.

Do not expose Demeter directly to the internet. Options for safer access:

- Keep it on a trusted LAN or VLAN.
- Bind it to a private interface, for example
  `demeter server --address 192.168.1.20:8249`.
- Put it behind a reverse proxy that enforces TLS and authentication.
- Use a firewall or VPN for remote access.

If you add a reverse proxy, remember that the sensor WebSocket also needs to
be proxied, including the `Upgrade` and `Connection` headers.

## Request conventions

- Send JSON request bodies with `Content-Type: application/json`.
- IDs are JSON numbers and database `int64` values.
- Successful creates return `201 Created` with the created object.
- Successful deletes return `204 No Content` with an empty body. Deleting an
  ID that does not exist also returns `204`; it is not reported as `404`.
- Validation and not-found errors return a JSON object with an `error`
  string:

```json
{ "error": "enclosure not found" }
```

Some internal `500` responses are written by Gin's error helper and may have
an empty or plain-text body, so do not assume every error response is JSON.

- There is no pagination. List endpoints return every matching row except
  where a `limit` parameter is documented.
- Most list endpoints return JSON **`null`**, not `[]`, when there are no
  rows. Metric-series endpoints initialize a slice and return `[]`. Clients
  should treat both as an empty result.
- Names of enclosures, systems, flows, genera, and species are unique. The
  database rejects duplicates, but the HTTP status varies: genus and species
  handlers return `409 Conflict` with a JSON error, while enclosure, system,
  and flow handlers currently return `500` with an empty body because they
  check the primary SQLite constraint code instead of the extended code.

### Referential integrity is not enforced

The schema declares `REFERENCES` clauses, but the current server opens SQLite
without enabling foreign-key enforcement (`PRAGMA foreign_keys = ON`). SQLite
therefore accepts rows whose parent ID does not exist.

This means the API can create a system with a missing enclosure, a flow with a
missing system or parent flow, a plant site with a missing flow, a species with
a missing genus, a plant with a missing site or species, a note with a missing
parent, or a sensor reading with a missing target. Those rows are stored, but
they are not associated with any real entity. Most are invisible in the UI and
in joined queries, although the flat plant-note card can still display an
orphan plant note.

Always create parents first and copy their IDs carefully. Required fields and
unique names are still validated; missing **parents** are not. The `409`
responses from the genus and species routes are for duplicate names, not for a
bad parent ID. Duplicate enclosure, system, and flow names currently surface
as an empty `500` instead.

## Response field casing

Responses are Go structs encoded with the standard library. Fields use
**PascalCase**, not snake_case:

```json
{
  "ID": 1,
  "Name": "Vegetable Tent"
}
```

Requests, by contrast, use lower-case JSON names such as `"name"` and
`"enclosure_id"`. Do not assume a response field is named like the request
field.

The timelapse list is an exception: its response objects use a lower-case
`"name"` key, because that response type has an explicit JSON tag.

## Nullable fields are wrapper objects

The database layer uses Go's `database/sql` nullable types. They are encoded
as objects, not as `null` or plain values:

| Type | JSON shape | Example |
| --- | --- | --- |
| `sql.NullInt64` | `{"Int64": <number>, "Valid": <bool>}` | `{"Int64":0,"Valid":false}` |
| `sql.NullString` | `{"String": <string>, "Valid": <bool>}` | `{"String":"","Valid":false}` |
| `sql.NullTime` | `{"Time": <RFC3339 string>, "Valid": <bool>}` | `{"Time":"0001-01-01T00:00:00Z","Valid":false}` |

For example, a flow with no parent looks like:

```json
{
  "ID": 3,
  "Name": "Left Channel",
  "SystemID": 1,
  "ParentFlowID": { "Int64": 0, "Valid": false }
}
```

When `Valid` is `false`, ignore the value. A plant with no species has
`"SpeciesID": {"Int64":0,"Valid":false}`; a note with no content has
`"Content": {"String":"","Valid":false}`.

Non-nullable `time.Time` fields, such as `PlantSiteNote.Timestamp`, are plain
RFC3339 strings.

## Core resources

### Enclosures

```sh
curl http://localhost:8249/api/enclosures
curl http://localhost:8249/api/enclosures/1
curl -X POST http://localhost:8249/api/enclosures \
  -H 'Content-Type: application/json' -d '{"name":"Vegetable Tent"}'
curl -X PUT http://localhost:8249/api/enclosures/1 \
  -H 'Content-Type: application/json' -d '{"name":"Veg Tent"}'
curl -X DELETE http://localhost:8249/api/enclosures/1
```

Response:

```json
{ "ID": 1, "Name": "Vegetable Tent" }
```

### Systems

```sh
curl http://localhost:8249/api/systems
curl 'http://localhost:8249/api/systems?enclosure_id=1'
curl -X POST http://localhost:8249/api/systems \
  -H 'Content-Type: application/json' \
  -d '{"name":"Main Reservoir","enclosure_id":1}'
```

Response:

```json
{ "ID": 1, "Name": "Main Reservoir", "EnclosureID": 1 }
```

### Flows

```sh
curl http://localhost:8249/api/flows
curl 'http://localhost:8249/api/flows?system_id=1'
curl 'http://localhost:8249/api/flows?parent_flow_id=2'
curl -X POST http://localhost:8249/api/flows \
  -H 'Content-Type: application/json' \
  -d '{"name":"Left Channel","system_id":1,"parent_flow_id":null}'
```

Response:

```json
{
  "ID": 3,
  "Name": "Left Channel",
  "SystemID": 1,
  "ParentFlowID": { "Int64": 0, "Valid": false }
}
```

If both `system_id` and `parent_flow_id` are supplied to the list endpoint,
`system_id` wins.

### Plant sites

```sh
curl http://localhost:8249/api/plant-sites
curl 'http://localhost:8249/api/plant-sites?flow_id=3'
curl -X POST http://localhost:8249/api/plant-sites \
  -H 'Content-Type: application/json' \
  -d '{"flow_id":3,"x":1,"y":0,"z":2}'
```

Response:

```json
{ "ID": 7, "FlowID": 3, "X": 1, "Y": 0, "Z": 2 }
```

`x`, `y`, and `z` are required on create and update, including when the value
is zero.

### Plants

```sh
curl http://localhost:8249/api/plants
curl 'http://localhost:8249/api/plants?plant_site_id=7'
curl 'http://localhost:8249/api/plants?species_id=3'
curl -X POST http://localhost:8249/api/plants \
  -H 'Content-Type: application/json' \
  -d '{"plant_site_id":7,"planted_on":"2026-04-12T09:30:00Z","species_id":3}'
```

Response:

```json
{
  "ID": 5,
  "PlantSiteID": 7,
  "PlantedOn": { "Time": "2026-04-12T09:30:00Z", "Valid": true },
  "SpeciesID": { "Int64": 3, "Valid": true }
}
```

If both filters are supplied, `plant_site_id` wins.

### Genera and species

```sh
curl http://localhost:8249/api/plant-genera
curl -X POST http://localhost:8249/api/plant-genera \
  -H 'Content-Type: application/json' -d '{"name":"Lactuca"}'

curl http://localhost:8249/api/plant-species
curl 'http://localhost:8249/api/plant-species?plant_genus_id=1'
curl -X POST http://localhost:8249/api/plant-species \
  -H 'Content-Type: application/json' \
  -d '{"name":"sativa","plant_genus_id":1}'
```

Response shapes:

```json
{ "ID": 1, "Name": "Lactuca" }
```

```json
{
  "ID": 3,
  "Name": "sativa",
  "PlantGenusID": { "Int64": 1, "Valid": true }
}
```

The generated Go type for a species is `PlantSpecy`; the JSON shape is still
the fields shown above.

## Sensor readings

Raw-reading endpoints return **newest first**. They accept `limit` (default
`100`) on enclosures and systems. Invalid or non-positive limits fall back to
the default rather than returning an error.

```sh
curl 'http://localhost:8249/api/enclosures/1/air-temperature?limit=10'
curl 'http://localhost:8249/api/enclosures/1/air-humidity?limit=10'
curl 'http://localhost:8249/api/systems/1/ph?limit=10'
curl 'http://localhost:8249/api/systems/1/ec?limit=10'
curl 'http://localhost:8249/api/systems/1/water-temperature?limit=10'
curl 'http://localhost:8249/api/plant-sites/7/lux'
```

Response (enclosure temperature):

```json
[
  {
    "ID": 941,
    "Timestamp": { "Time": "2026-11-03T14:05:12Z", "Valid": true },
    "EnclosureID": 1,
    "TemperatureC": 24.512
  }
]
```

Values are stored as floating-point numbers in their natural units: °C,
percent RH, pH, EC, and lux. The `× 1000` encoding applies only on the sensor
WebSocket wire, not to the REST representation.

There are no read endpoints for dissolved oxygen, water proximity, or flow
rate in the current API.

## Metric series

Metric endpoints return interval averages intended for charting. The current
web panel uses them for enclosure air charts; system and plant charts still use
raw-reading endpoints:

```text
GET /api/enclosures/{id}/metrics/air-temperature
GET /api/enclosures/{id}/metrics/air-humidity
GET /api/systems/{id}/metrics/ph
GET /api/systems/{id}/metrics/ec
GET /api/systems/{id}/metrics/water-temperature
GET /api/plant-sites/{id}/metrics/lux
```

Query parameters:

| Parameter | Default | Meaning |
| --- | --- | --- |
| `start` | 24 hours ago | RFC3339 start of the window. |
| `end` | now | RFC3339 end of the window. |
| `limit` | `24` | Number of equal-length intervals, 1–10000. |

The server divides `[start, end)` into `limit` equal intervals, averages all
readings in each interval, and returns one point per non-empty interval.
Intervals are emitted in chronological order. Empty intervals are omitted, so
the array may contain fewer than `limit` points. Each point's `Timestamp` is
the **start** of its interval, not the timestamp of a raw reading.

Example:

```sh
curl 'http://localhost:8249/api/systems/1/metrics/ph?start=2026-11-02T00:00:00Z&end=2026-11-03T00:00:00Z&limit=24'
```

```json
[
  { "Timestamp": "2026-11-02T00:00:00Z", "Value": 6.41 },
  { "Timestamp": "2026-11-02T01:00:00Z", "Value": 6.44 },
  { "Timestamp": "2026-11-02T03:00:00Z", "Value": 6.39 }
]
```

The third interval (`02:00`) had no readings and is omitted.

Errors:

- `400` for an invalid ID, an unparseable `start` or `end`, `start >= end`, or
  a `limit` outside 1–10000.
- `500` for a database failure.

The enclosure charts' range buttons choose `start`, `end`, and `limit`; there is
no separate aggregation interval control.

## Topology helpers

These endpoints return joined data for a parent entity:

```sh
curl http://localhost:8249/api/enclosures/1/plants
curl http://localhost:8249/api/flows/3/plants
```

`GET /api/enclosures/{id}/plants` returns plants in the enclosure's systems
and flows. `GET /api/flows/{id}/plants` returns plants on the flow's sites.
Both use inner joins to `plant_species` and `plant_genus`, so they omit a plant
with no species and a species with no genus. The plain `GET /api/plants` list
does not have that limitation.

Both routes ignore a `limit` query parameter. The flow helper returns a joined
shape with nested coordinates:

```json
[
  {
    "ID": 5,
    "PlantedOn": { "Time": "2026-04-12T09:30:00Z", "Valid": true },
    "PlantSiteID": 7,
    "X": { "Int64": 1, "Valid": true },
    "Y": { "Int64": 0, "Valid": true },
    "Z": { "Int64": 2, "Valid": true },
    "FlowName": "Left Channel",
    "FlowID": 3,
    "SpeciesID": 3,
    "SpeciesName": "sativa",
    "GenusName": "Lactuca"
  }
]
```

## Notes

Simple notes (enclosure, system, flow) share one request shape:

```json
{ "timestamp": "2026-11-03T14:05:12Z", "content": "Changed nutrients." }
```

Both fields are optional on create; an omitted timestamp becomes the current
server time. `PUT` is a replacement, not a partial merge: send both fields you
want to keep. An omitted content becomes SQL `NULL`, and an omitted timestamp
becomes the current server time. Responses use `NullTime` and `NullString`
wrappers.

```sh
curl http://localhost:8249/api/enclosure-notes
curl -X POST http://localhost:8249/api/system-notes \
  -H 'Content-Type: application/json' -d '{"content":"Topped up 10 L."}'
```

Enclosure, system, and flow notes are **not linked to their parent**. Their
tables have no parent ID, and every list endpoint returns all notes of that
type. See [Notes](../notes#current-scoping-limitation).

Plant-site notes require `plant_site_id`:

```json
{ "plant_site_id": 7, "content": "Lower leaves yellowing." }
```

Plant notes use the misspelled field **`plan_id`** for the plant ID:

```json
{ "plan_id": 5, "content": "First true leaves." }
```

The field is `plan_id` in both request bodies and query strings, and
`PlanID` in responses. The current server does not enforce the note parent
foreign keys, so a nonexistent `plant_site_id` or `plan_id` is accepted and the
note becomes an orphan.

## Timelapse

```sh
# List entries under the timelapse root.
curl http://localhost:8249/api/timelapse

# Stream frames (binary MJPEG; use a player or browser).
curl http://localhost:8249/api/timelapse/a1b2c3d4e5f6/timelapse.mjpeg
```

List response:

```json
[
  { "name": "a1b2c3d4e5f6" },
  { "name": "f6e5d4c3b2a1" }
]
```

The list handler reads the fixed root `/srv/demeter/timelapse`. If that root is
missing or unreadable it aborts with `500` but still writes an empty `[]` body,
so check the status code, not only the JSON. The stream handler returns `500`
if the MAC directory cannot be read.

The stream is assembled on demand from `.jpeg` files, flipped vertically, and
served as `multipart/x-mixed-replace`. It is not a stored video. See
[Photos and Timelapse](../photos-and-timelapse).

## Endpoint index

### Core objects

| Method | Path |
| --- | --- |
| `GET` | `/api/enclosures` |
| `POST` | `/api/enclosures` |
| `GET` `PUT` `DELETE` | `/api/enclosures/{id}` |
| `GET` | `/api/enclosures/{id}/plants` |
| `GET` | `/api/systems` (`?enclosure_id=`) |
| `POST` | `/api/systems` |
| `GET` `PUT` `DELETE` | `/api/systems/{id}` |
| `GET` | `/api/flows` (`?system_id=` or `?parent_flow_id=`) |
| `POST` | `/api/flows` |
| `GET` `PUT` `DELETE` | `/api/flows/{id}` |
| `GET` | `/api/flows/{id}/plants` |
| `GET` | `/api/plant-sites` (`?flow_id=`) |
| `POST` | `/api/plant-sites` |
| `GET` `PUT` `DELETE` | `/api/plant-sites/{id}` |
| `GET` | `/api/plants` (`?plant_site_id=` or `?species_id=`) |
| `POST` | `/api/plants` |
| `GET` `PUT` `DELETE` | `/api/plants/{id}` |
| `GET` | `/api/plant-genera` |
| `POST` | `/api/plant-genera` |
| `GET` `PUT` `DELETE` | `/api/plant-genera/{id}` |
| `GET` | `/api/plant-species` (`?plant_genus_id=`) |
| `POST` | `/api/plant-species` |
| `GET` `PUT` `DELETE` | `/api/plant-species/{id}` |

### Readings and metrics

| Method | Path |
| --- | --- |
| `GET` | `/api/enclosures/{id}/air-temperature` (`?limit=`) |
| `GET` | `/api/enclosures/{id}/air-humidity` (`?limit=`) |
| `GET` | `/api/enclosures/{id}/metrics/air-temperature` |
| `GET` | `/api/enclosures/{id}/metrics/air-humidity` |
| `GET` | `/api/systems/{id}/ph` (`?limit=`) |
| `GET` | `/api/systems/{id}/ec` (`?limit=`) |
| `GET` | `/api/systems/{id}/water-temperature` (`?limit=`) |
| `GET` | `/api/systems/{id}/metrics/ph` |
| `GET` | `/api/systems/{id}/metrics/ec` |
| `GET` | `/api/systems/{id}/metrics/water-temperature` |
| `GET` | `/api/plant-sites/{id}/lux` |
| `GET` | `/api/plant-sites/{id}/metrics/lux` |

### Notes

| Method | Path |
| --- | --- |
| `GET` `POST` | `/api/enclosure-notes` |
| `GET` `PUT` `DELETE` | `/api/enclosure-notes/{id}` |
| `GET` `POST` | `/api/system-notes` |
| `GET` `PUT` `DELETE` | `/api/system-notes/{id}` |
| `GET` `POST` | `/api/flow-notes` |
| `GET` `PUT` `DELETE` | `/api/flow-notes/{id}` |
| `GET` `POST` | `/api/plant-site-notes` (`?plant_site_id=`) |
| `GET` `PUT` `DELETE` | `/api/plant-site-notes/{id}` |
| `GET` `POST` | `/api/plant-notes` (`?plan_id=`) |
| `GET` `PUT` `DELETE` | `/api/plant-notes/{id}` |

### Sensors and media

| Method | Path |
| --- | --- |
| `GET` (WebSocket) | `/api/websocket/sensors` |
| `GET` | `/api/timelapse` |
| `GET` | `/api/timelapse/{mac}/timelapse.mjpeg` |

## OpenAPI

The repository's OpenAPI document, `docs/static/openapi.yaml`, lists the
routes and schemas and generates an interactive reference. It describes the
intended contract; where it differs from observed behaviour in this guide
(such as duplicate-name status codes or empty-list `null` values), trust the
runtime notes here. Build the interactive reference with:

```sh
cd docs
task openapi
```

Then open `http://localhost:1313/openapi.html` while the docs server is
running. See [Building the Documentation](../operating-and-maintaining-demeter#building-the-documentation)
for the full docs workflow.
