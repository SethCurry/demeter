---
title: "Flows and Plant Sites"
summary: "Build water loops, nest child loops, lay out plant sites in three dimensions, and understand how flow rate fits in."
date: 2023-09-07T16:13:18+02:00
lastmod: 2026-11-03T00:00:00+00:00
draft: false
weight: 7
toc: true
---

A **flow** is one path that water takes through a system. A **plant site** is a
single position within a flow where a plant can grow. Together they describe
the physical layout of a hydroponic build.

<figure>
  <img src="/images/sample.jpg" alt="Flow detail page with 3D plant site layout" width="1200" height="675" loading="lazy">
  <figcaption>Flow detail combines child loops, plant sites, plants, and notes. Replace this placeholder with a screenshot.</figcaption>
</figure>

## The flow hierarchy

```text
Enclosure
└── System
    ├── Flow (root)
    │   ├── Plant Site (0, 0, 0)
    │   ├── Plant Site (1, 0, 0)
    │   └── Child Flow
    │       └── Plant Site (0, 0, 1)
    └── Flow (another root)
```

A flow always belongs to a **system**. It may also have a **parent flow**. The
parent link lets you model a main loop that splits into smaller loops, such as
a manifold feeding several grow channels. Parent and child flows are expected
to belong to the same system, but the current database does not enforce that;
the UI's parent picker shows every flow. Choose carefully.

Flows own:

- **Plant sites** located directly on that loop.
- **Child flows**.
- **Flow rate** readings (accepted and stored, but not exposed by the API yet).
- **Flow notes**.

Plant sites do **not** inherit each other's readings, and a flow's readings are
not summed up from its children. Each entity is addressed independently by ID.

## Create a flow

1. Open **Hydroponics &rarr; Flows**.
2. Click **New Flow**.
3. Enter a unique **Name**.
4. Choose the **System** that supplies the loop.
5. Optionally choose a **Parent Flow**. Leave it empty for a root flow.
6. Click **Create**.

Names are unique across all flows. The database rejects a duplicate, but the
current flow create/update handlers return an empty `500` response rather than
the intended `409`. Choose an existing system and parent flow. The current
server does not enable SQLite foreign-key enforcement, so the API can store a
flow whose `system_id` or `parent_flow_id` does not exist. Those rows will not
appear under a real system or parent, so use the UI pickers or verify IDs
before sending a request.

API example:

```sh
curl -X POST http://localhost:8249/api/flows \
  -H 'Content-Type: application/json' \
  -d '{"name":"Left Channel","system_id":1,"parent_flow_id":2}'
```

Omit `parent_flow_id`, or send `null`, for a root flow.

## The flow list

The table shows each flow's:

- **ID** and **Name**.
- **System**.
- **Parent Flow**, or a `root` tag.
- Number of **Plant Sites**.

Use the list to see the shape of the installation at a glance. The API can
also filter flows with `GET /api/flows?system_id=1` or
`GET /api/flows?parent_flow_id=2`.

## The flow detail page

The detail page shows:

- The flow's ID, name, parent flow, system, and enclosure, with links to each.
- Counts of **child flows**, **plant sites**, and **plants**.
- A table of child flows with their ID, name, and plant-site count.
- A **Plant Site Layout (3D)** chart of this flow's sites.
- A list of plants on this flow's sites.
- A notes card.

### The 3D layout

The 3D view is a Plotly scatter plot. Coordinates follow the database
conventions:

| Axis | Direction |
| --- | --- |
| **X** | left to right |
| **Y** | bottom to top |
| **Z** | back to front |

The flow detail page plots the sites that belong to the current flow. A marker
is filled in the flow's colour when at least one plant occupies the site, and
is grey when the site is empty. Hover a marker to see the site ID, its true
coordinates, the flow name, and the plants on it. Coordinates are multiplied
by four for display so adjacent sites remain visually distinct; the hover text
always shows the real database values.

Occupancy and the plants list come from `GET /api/flows/{id}/plants`. That
query inner-joins `plant_species` and `plant_genus`, so a plant with no species,
or a species with no genus, is invisible on the flow page: its site looks empty
and the flow's plant count excludes it. The **Plants** list and plant detail
page still show it. Classify the plant to include it in the flow view.

The plant detail page uses the same component in highlight mode: the selected
plant's site stands out and all other sites are grey.

## Create a plant site

Plant sites are created from the **Plants** page, because a site and its
occupant are managed together:

1. Open **Hydroponics &rarr; Plants**.
2. Click **New Plant Site**.
3. Choose the **Flow**.
4. Enter integer **X**, **Y**, and **Z** coordinates. They default to `0`.
5. Click **Create**.

Choose a flow that already exists. Coordinates are integers; negative values
are allowed by the database but may look odd in the 3D view. The current server
does not enforce the flow foreign key, so the API can also store a plant site
whose `flow_id` does not exist; it will not appear under any real flow. Choose a
coordinate system that matches your physical build and stay consistent,
because the same axes are used on every flow and plant page.

API example:

```sh
curl -X POST http://localhost:8249/api/plant-sites \
  -H 'Content-Type: application/json' \
  -d '{"flow_id":3,"x":1,"y":0,"z":2}'
```

The response contains the new plant site `ID`, which is also the target used
by a BH1750 light sensor.

## Plant sites on the Plants page

The **Plants** page is a table of plant sites, not a table of plants. Each row
shows:

- **Site ID**
- **Flow**
- **X**, **Y**, **Z**
- The **Plant** currently assigned to the site, or an `empty` tag
- **Species**
- **Planted** date

Clicking a plant ID opens the plant detail page. A site can technically hold
multiple plants through the API, but the table displays only one plant per
site. The flow detail 3D hover text can list several, subject to the
classification caveat above.

## Flow rate

The server accepts flow-rate measurements as message type `9` and stores them
in the `flow_rate` table in metres per second. The current REST API has no
flow-rate read endpoint, and the web panel has no flow-rate card or chart. A
custom sensor can send the readings now; they will become visible when a read
endpoint is added.

```text
9:<flow_id>:<rate_mps × 1000>
```

For example, `9:3:1250` records 1.25 m/s for flow `3`.

## Rename or delete flows and plant sites

Use the API while the UI is view/create-only:

```sh
# Flow
curl -X PUT http://localhost:8249/api/flows/3 \
  -H 'Content-Type: application/json' \
  -d '{"name":"Right Channel","system_id":1,"parent_flow_id":null}'

curl -X DELETE http://localhost:8249/api/flows/3

# Plant site
curl -X PUT http://localhost:8249/api/plant-sites/7 \
  -H 'Content-Type: application/json' \
  -d '{"flow_id":3,"x":1,"y":0,"z":3}'

curl -X DELETE http://localhost:8249/api/plant-sites/7
```

Deletion does not cascade to readings, notes, or plants. Because foreign keys
are not enforced, the delete succeeds even when dependants exist and leaves
dangling references. Remove dependants first and back up the database.

## API endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/flows` | List flows; optional `system_id` or `parent_flow_id` filter. |
| `POST` | `/api/flows` | Create a flow. |
| `GET` | `/api/flows/{id}` | Get one flow. |
| `PUT` | `/api/flows/{id}` | Update a flow. |
| `DELETE` | `/api/flows/{id}` | Delete a flow. |
| `GET` | `/api/flows/{id}/plants` | List plants on the flow's sites. |
| `GET` | `/api/plant-sites` | List plant sites; optional `flow_id` filter. |
| `POST` | `/api/plant-sites` | Create a plant site. |
| `GET` | `/api/plant-sites/{id}` | Get one plant site. |
| `PUT` | `/api/plant-sites/{id}` | Update a plant site. |
| `DELETE` | `/api/plant-sites/{id}` | Delete a plant site. |
| `GET` | `/api/plant-sites/{id}/lux` | Latest raw lux readings. |
| `GET` | `/api/plant-sites/{id}/metrics/lux` | Averaged lux series. |

See [REST API](../rest-api) and
[Plants and Taxonomy](../plants-and-taxonomy).
