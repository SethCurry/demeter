---
title: "Notes"
summary: "Record observations against enclosures, systems, flows, plant sites, and plants, including timestamps and current scoping limitations."
date: 2023-09-07T16:13:18+02:00
lastmod: 2026-11-03T00:00:00+00:00
draft: false
weight: 9
toc: true
---

Notes are timestamped free text. They are meant for observations that do not
fit a numeric metric: a reservoir top-up, a pest sighting, a nutrient change,
or a change in lighting schedule.

## Where notes appear

The web panel lets you add and read notes on three detail pages:

| Page | Button | Note type |
| --- | --- | --- |
| Enclosure detail | **Add Note** | Enclosure note |
| System detail | **Add Note** | System note |
| Flow detail | **Add Note** | Flow note |

Plant detail pages display a **Plant Notes** card, but the UI has no button to
add one and does not filter the list: every plant detail page shows every plant
note. Plant-site notes are available through the API only. See
[Adding notes through the API](#adding-notes-through-the-api).

<figure>
  <img src="/images/sample.jpg" alt="Notes card on a detail page" width="1200" height="675" loading="lazy">
  <figcaption>Enclosure, system, and flow detail pages have a notes card with an Add Note button. Replace this placeholder with a screenshot.</figcaption>
</figure>

## Add a note in the UI

1. Open an enclosure, system, or flow detail page.
2. Click **Add Note**.
3. Optionally set a **Timestamp**. If you leave it empty, the server records
   the current time.
4. Enter the **Content**. The UI requires text.
5. Submit the form.

The notes table refreshes after the note is created. Notes are sorted by the
order returned by the API, not by any client-side sort.

## How timestamps work

For enclosure, system, and flow notes, the request body accepts an optional
RFC3339 timestamp:

```json
{
  "timestamp": "2026-04-12T09:30:00Z",
  "content": "Changed nutrient solution; EC target is now 1.4."
}
```

If `timestamp` is omitted or `null`, the server substitutes the current time.
This is true even though the underlying database column also has a
`CURRENT_TIMESTAMP` default: the API always supplies a value.

Plant-site and plant notes use a non-null timestamp column. Their handlers
also default an omitted timestamp to the current time.

## Current scoping limitation

The `enclosure_note`, `system_note`, and `flow_note` tables do not have a
foreign-key column pointing at their parent. They are standalone rows with an
ID, timestamp, and content. The API therefore exposes a flat list per note
type, and every detail page of that type displays the same list.

In practice:

- A note added from **Enclosure A** also appears on **Enclosure B**.
- A note added from **System A** appears on every system detail page.
- A note added from **Flow A** appears on every flow detail page.
- Deleting the parent does not delete or orphan-link its notes.

Plant-site and plant notes carry parent IDs in `plant_site_id` and `plan_id`,
but the UI does not add or filter by them. The database declares these as
foreign keys, yet the current server does not enable SQLite foreign-key
enforcement, so a nonexistent parent is also accepted. A plant-site note with a
bad parent is not shown anywhere in the UI. A plant note with a bad `plan_id`
still appears in the unfiltered **Plant Notes** card on every plant detail
page.

## Adding notes through the API

### Enclosure, system, and flow notes

The three simple note types share a request shape:

```sh
# Enclosure note
curl -X POST http://localhost:8249/api/enclosure-notes \
  -H 'Content-Type: application/json' \
  -d '{"content":"Added a second fan."}'

# System note
curl -X POST http://localhost:8249/api/system-notes \
  -H 'Content-Type: application/json' \
  -d '{"timestamp":"2026-04-12T09:30:00Z","content":"Topped up 10 L."}'

# Flow note
curl -X POST http://localhost:8249/api/flow-notes \
  -H 'Content-Type: application/json' \
  -d '{"content":"Replaced the pump."}'
```

Each type supports list, get, update, and delete:

```text
GET    /api/enclosure-notes
POST   /api/enclosure-notes
GET    /api/enclosure-notes/{id}
PUT    /api/enclosure-notes/{id}
DELETE /api/enclosure-notes/{id}
```

The same pattern exists under `/api/system-notes` and `/api/flow-notes`.
`PUT` replaces the timestamp and content; send both fields.

### Plant-site notes

Plant-site notes require a `plant_site_id` and can be filtered by it:

```sh
curl -X POST http://localhost:8249/api/plant-site-notes \
  -H 'Content-Type: application/json' \
  -d '{"plant_site_id":7,"content":"Lower leaves yellowing."}'

curl 'http://localhost:8249/api/plant-site-notes?plant_site_id=7'
```

### Plant notes

Plant notes use the field name **`plan_id`** in the request and query string
(not `plant_id`). This is the name used by the current schema and API:

```sh
curl -X POST http://localhost:8249/api/plant-notes \
  -H 'Content-Type: application/json' \
  -d '{"plan_id":5,"content":"First true leaves appeared."}'

curl 'http://localhost:8249/api/plant-notes?plan_id=5'
```

If you are writing a client, treat `plan_id` as the plant identifier until the
field is renamed.

## Notes and Home Assistant

The Home Assistant integration does not currently import notes. Notes are
available only through the Demeter web panel and REST API.

## API reference

For the complete list of note endpoints and response shapes, see
[REST API](../rest-api) and the
[OpenAPI reference](/openapi.html).
