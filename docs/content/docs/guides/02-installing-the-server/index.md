---
title: "Installing the Server"
summary: "Build the Demeter server from source and run it, either directly on your host or inside the Docker dev stack."
date: 2023-09-07T16:13:18+02:00
lastmod: 2023-09-07T16:13:18+02:00
draft: false
weight: 2
toc: true
---

The Demeter server is a single Go binary that embeds the web control panel
and the SQLite schema, so installing it is mostly a matter of building the
binary and running it. This guide walks through both a host install and the
containerised dev stack, and ends with a quick check that the server is up.

## What you are installing

The server (`cmd/demeter`) does three things at runtime:

1. Listens on an HTTP address (default `:8249`) and serves the REST API and
   embedded web panel.
2. Accepts WebSocket connections from ESP32 sensors at
   `/api/websocket/sensors`.
3. Reads and writes an embedded SQLite database (`./demeter.sqlite` by
   default), creating the schema on first run.

It has no external runtime dependencies beyond the binary and a place to keep
its SQLite file.

## Prerequisites

To build from source you need:

- **Go** 1.27 or newer; see the [`go` directive in
  `go.mod`](https://go.dev/dl/).
- **Node.js** and npm. The web control panel is a React + TypeScript
  app that is built and embedded into the binary.
- **[Task](https://taskfile.dev)** the build orchestration tool. The
  repo's `Taskfile.yml` drives the web build, copies the SQL schema and UI
  assets into `internal/embedded/`, and then runs `go build`.
- **Git** to clone the repository.

The containerised option additionally needs **Docker with the Compose
plugin**.

## Build from source

Clone the repository and build the server. The `build` task handles
everything: it builds the web UI, copies `sql/schema.sql` and the built web
assets into `internal/embedded/`, and then compiles `cmd/demeter`.

```sh
git clone https://github.com/SethCurry/demeter.git
cd demeter
task build
```

When it finishes you will have a `./demeter` binary at the repository root.
Because the UI and schema are embedded, that single binary is the whole
server &mdash; nothing else needs to be deployed alongside it.

> If you want to skip Task and drive the steps yourself, the `build` task is
> equivalent to: `task ui:build`, then copying `sql/schema.sql` to
> `internal/embedded/schema.sql`, replacing `internal/embedded/ui` with
> `web/build`, and finally `go build ./cmd/demeter`.

## Run on the host

Start the server with the `server` command:

```sh
./demeter server
```

By default it listens on `:8249` and writes `./demeter.sqlite` in the current
working directory. Open the control panel at
<http://localhost:8249>.

### Choosing a different address

Pass `--address` to listen somewhere other than `:8249`:

```sh
./demeter server --address :9000
./demeter server --address 127.0.0.1:8249
```

The sensor simulator and any client (the web panel, Home Assistant) just need
to point at the same address.

### Where the data lives

The SQLite database path is currently fixed to `./demeter.sqlite` relative to
the server's working directory, so run the binary from whatever directory you
want the database to live in. On first start the server applies the embedded
schema automatically.

## Run in Docker

The repository ships a development Docker Compose stack under `dev/` that
runs the real server alongside Home Assistant. It is the easiest way to get a
full stack running without installing Go or Node.

Build the binary first (Task is still required for the one-time build), then
start the stack:

```sh
task build
./dev/serve.sh
```

`serve.sh` runs the Compose `live` profile, bind-mounts the binary built at
the repository root, keeps the SQLite database on the `demeter-data` Docker
volume, and prints the URLs for Home Assistant and the Demeter server when it
is done:

```text
Home Assistant: http://localhost:8123
demeter server: http://localhost:8249
```

A few environment variables are useful when running the containerised server:

| Variable               | Default       | Purpose                                            |
| ---------------------- | ------------- | -------------------------------------------------- |
| `DEMETER_BINARY`       | `../demeter`  | Binary bind-mounted into the server container      |
| `DEMETER_SERVER_PORT`  | `8249`        | Host port published for the Demeter server         |
| `HA_PORT`              | `8123`        | Host port for Home Assistant                       |

Stop the stack with:

```sh
./dev/down.sh
```

See `dev/README.md` for the full set of options, including how to point a
host-side Home Assistant container at a server running on the host via
`http://host.docker.internal:8249`.

## Verify the server is running

Once the server is up &mdash; on the host or in Docker &mdash; check the REST
API. The enclosures endpoint should return JSON (an empty array on a fresh
database):

```sh
curl http://localhost:8249/api/enclosures
```

Then load the control panel in a browser:

```text
http://localhost:8249
```

## Feed it some data

With the server running you can exercise the whole pipeline without any
hardware using the built-in sensor simulator. It opens a WebSocket to the
server and streams synthetic readings every few seconds:

```sh
./demeter simulate-sensor --addr localhost:8249
```

In the Docker stack the simulator is available inside the server container:

```sh
docker compose -f dev/docker-compose.yml exec demeter \
    /usr/local/bin/demeter simulate-sensor --addr demeter:8249
```

Leave the simulator running and watch the readings appear in the control
panel. When you are ready to move to real hardware, continue with the sensor
and controller reference docs to flash an ESP32.
