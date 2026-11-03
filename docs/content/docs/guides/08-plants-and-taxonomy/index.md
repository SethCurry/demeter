---
title: "Plants and Taxonomy"
summary: "Assign plants to sites, record planting dates, and organise them with genera and species."
description: "Assign plants to sites, record planting dates, classify specimens with genera and species, and read plant-site lux history."
date: 2023-09-07T16:13:18+02:00
lastmod: 2026-11-03T00:00:00+00:00
draft: false
weight: 8
toc: true
---

Plants are the reason the rest of the hierarchy exists. A **plant** is an
individual specimen assigned to a **plant site**, optionally classified with a
**species** and **genus**.

<figure>
  <img src="/images/sample.jpg" alt="Plant list and plant detail page" width="1200" height="675" loading="lazy">
  <figcaption>The Plants page combines plant sites with their current occupant. Replace this placeholder with a screenshot.</figcaption>
</figure>

## The taxonomy model

Demeter keeps taxonomy deliberately simple:

```text
Plant Genus (optional classification)
└── Plant Species (optional genus)
    └── Plant (optional species)
        └── Plant Site (required)
```

- A **genus** has a unique name.
- A **species** has a unique name and an optional genus.
- A **plant** may have a species, and always has a plant site.

Names are global, not scoped to a genus: two species in different genera
cannot share a name. The UI displays a species as `<Genus> <Species>` when a
genus is set, otherwise just the species name.

Taxonomy is descriptive only. A species does not carry care instructions,
expected pH ranges, or growth duration in the current schema.

## Create genera and species

### Genus

1. Open **Taxonomy &rarr; Genera**.
2. Click **New Genus**.
3. Enter a unique **Name**, for example `Lactuca`.
4. Click **Create**.

The list shows each genus, its ID, and how many species belong to it. Clicking
a genus opens its detail page, where you can see and add species.

### Species

1. Open **Taxonomy &rarr; Species**.
2. Click **New Species**.
3. Enter a unique **Name**, for example `sativa`.
4. Optionally choose a **Genus**.
5. Click **Create**.

You can also add a species directly from a genus detail page with **Add
Species**; the genus is preselected.

API examples:

```sh
curl -X POST http://localhost:8249/api/plant-genera \
  -H 'Content-Type: application/json' \
  -d '{"name":"Lactuca"}'

curl -X POST http://localhost:8249/api/plant-species \
  -H 'Content-Type: application/json' \
  -d '{"name":"sativa","plant_genus_id":1}'
```

## Add a plant to a site

1. Open **Hydroponics &rarr; Plants**.
2. Click **Add Plant**.
3. Choose a **Plant Site**. The picker shows the site ID, flow name, and
   coordinates.
4. Optionally choose a **Planted On** date and time. If you leave it empty,
   the server records the current time.
5. Optionally choose a **Species**.
6. Click **Create**.

A site can be empty. Adding a plant does not remove an earlier plant from the
same site, so the API can represent succession or multiple specimens. The
Plants table shows one plant per site; the 3D layout's hover text can list
every plant at a site, while each plant still has its own detail page.

> **Taxonomy affects some plant lists.** The enclosure and flow plant cards
> use joined endpoints that inner-join `plant_species` and `plant_genus`. A
> plant with no species, or a species with no genus, still appears on the
> **Plants** page and its own detail page, but not in enclosure/flow plant
> counts or the flow 3D layout. Assign a species whose genus is set if you want
> the plant visible everywhere.

API example:

```sh
curl -X POST http://localhost:8249/api/plants \
  -H 'Content-Type: application/json' \
  -d '{"plant_site_id":7,"planted_on":"2026-04-12T09:30:00Z","species_id":3}'
```

Both `planted_on` and `species_id` are optional. Omitting `planted_on` uses
the current time; omitting `species_id` leaves the plant unclassified.

Use IDs that already exist. The current server does not enable SQLite
foreign-key enforcement, so a request with a nonexistent `plant_site_id` or
`species_id` is accepted and stored as an orphan. The plant will not appear
under a real site or species. The UI pickers only offer existing records.

## Plant list and plant detail

The **Plants** page is site-centric. Each row represents a plant site and
shows its flow, coordinates, occupant, species, and planting date. This makes
it easy to spot empty sites.

The **plant detail** page shows:

- The plant ID and planting date.
- Species, with a link to the species page.
- Plant site, flow, system, and enclosure, with links at every level.
- A **Lux** chart for the plant's site.
- A **Site Position** statistic and counts of sites and plants in the system.
- A **Plant Location in System (3D)** chart with the occupied site
  highlighted.
- A **Plant Notes** table.

The lux chart reads the `plant_site_lux` table by plant site ID, so it shows
light at the location rather than light for the individual plant. See
[Flows and Plant Sites](../flows-and-plant-sites).

## Genus and species detail pages

A **genus detail** page shows the genus name, a count of its species and
plants, and a species table with each species' plant count. It provides
**Add Species**.

A **species detail** page shows the species name, its genus, and a count and
list of plants of that species. The species page has no add-plant button; add
plants from the **Plants** page.

## Light readings

A BH1750 on a plant site sends illuminance as message type `10`:

```text
10:<plant_site_id>:<lux>
```

Unlike the other scalar metrics, lux is sent as a plain integer and is not
multiplied by 1000. Set `BH_PLANT_SITE_ID` in the firmware configuration to the
site's numeric ID. See
[Connecting ESP32 Sensors](../esp32-hardware).

## Rename or delete

The panel has no edit or delete controls for taxonomy or plants. Use the API:

```sh
# Genus
curl -X PUT http://localhost:8249/api/plant-genera/1 \
  -H 'Content-Type: application/json' -d '{"name":"Lactuca"}'
curl -X DELETE http://localhost:8249/api/plant-genera/1

# Species
curl -X PUT http://localhost:8249/api/plant-species/3 \
  -H 'Content-Type: application/json' \
  -d '{"name":"sativa","plant_genus_id":1}'
curl -X DELETE http://localhost:8249/api/plant-species/3

# Plant
curl -X PUT http://localhost:8249/api/plants/5 \
  -H 'Content-Type: application/json' \
  -d '{"plant_site_id":7,"planted_on":"2026-04-12T09:30:00Z","species_id":3}'
curl -X DELETE http://localhost:8249/api/plants/5
```

Deleting a genus does not cascade to its species. Deleting a species or a
plant site does not cascade to plants. Because foreign keys are not enforced,
these deletes succeed even when dependants exist and leave dangling references.
Clean up dependants first and back up the database.

## API endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/plants` | List plants; optional `plant_site_id` or `species_id` filter. |
| `POST` | `/api/plants` | Add a plant. |
| `GET` | `/api/plants/{id}` | Get one plant. |
| `PUT` | `/api/plants/{id}` | Update a plant. |
| `DELETE` | `/api/plants/{id}` | Delete a plant. |
| `GET` | `/api/plant-genera` | List genera. |
| `POST` | `/api/plant-genera` | Create a genus. |
| `GET` | `/api/plant-genera/{id}` | Get a genus. |
| `PUT` | `/api/plant-genera/{id}` | Update a genus. |
| `DELETE` | `/api/plant-genera/{id}` | Delete a genus. |
| `GET` | `/api/plant-species` | List species; optional `plant_genus_id` filter. |
| `POST` | `/api/plant-species` | Create a species. |
| `GET` | `/api/plant-species/{id}` | Get a species. |
| `PUT` | `/api/plant-species/{id}` | Update a species. |
| `DELETE` | `/api/plant-species/{id}` | Delete a species. |

See [REST API](../rest-api) for response shapes and note endpoints.
