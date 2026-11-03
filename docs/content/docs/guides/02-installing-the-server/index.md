---
title: "Installing the Server"
summary: "Build the Demeter server from source, run it on a host or in the Docker development stack, and verify that the API and web panel are live."
date: 2023-09-07T16:13:18+02:00
lastmod: 2026-11-03T00:00:00+00:00
draft: false
weight: 2
toc: true
---

The Demeter server is a single Go binary that embeds the web control panel and
the SQLite schema. Installing it means building that binary, choosing a
working directory, and running it.

<figure>
  <img src="/images/sample.jpg" alt="Demeter dashboard after installation" width="1200" height="675" loading="lazy">
  <figcaption>A fresh server serves the web panel at port 8249. Replace this placeholder with a screenshot.</figcaption>
</figure>

## What you are installing

The server (`cmd/demeter`) does four things at runtime:

1. Listens on an HTTP address (default `:8249`) and serves the REST API and
   embedded web panel.
2. Accepts WebSocket connections from sensors at `/api/websocket/sensors`.
3. Reads and writes a SQLite database at `./demeter.sqlite`, relative to its
   working directory, creating the tables when it starts against an empty
   database.
4. Stores timelapse frames under `/srv/demeter/timelapse`.

There is no configuration file and no authentication. See
[Security](#security-before-you-start).

## Prerequisites

To build from source you need:

- **Go 1.27.1 or newer**, matching `go.mod`.
- **Node.js and npm** for the React web panel. Use a current LTS release.
- **[Task](https://taskfile.dev)** for the build pipeline.
- **Git** to clone the repository.

The Docker path also needs **Docker with the Compose plugin**.

## Build from source

```sh
git clone https://github.com/SethCurry/demeter.git
cd demeter
task build
```

`task build` runs `task embed`, which:

1. Runs the web UI build (`task ui:build`).
2. Copies `sql/schema.sql` to `internal/embedded/schema.sql`.
3. Replaces `internal/embedded/ui` with `web/build`.
4. Runs `go build ./cmd/demeter`.

The result is `./demeter` at the repository root. The binary is self-contained:
the UI and schema are embedded.

> A plain `go build ./cmd/demeter` compiles whatever is already in
> `internal/embedded`. It does **not** rebuild the UI or refresh the schema
> copy. Use `task build` for a complete build.

## Run on the host

```sh
./demeter server
```

Defaults:

| Setting | Default |
| --- | --- |
| Listen address | `:8249` (all interfaces) |
| Database | `./demeter.sqlite` in the current working directory |
| Timelapse root | `/srv/demeter/timelapse` |

Open <http://localhost:8249>.

### Choose the listen address

```sh
./demeter server --address :9000
./demeter server --address 127.0.0.1:8249
./demeter server --address 192.168.1.20:8249
```

The flag is `--address` for the `server` command. The separate
`simulate-sensor` command uses `--addr` for the server it connects to.

### Choose the data directory

The database path is fixed to `./demeter.sqlite`, so choose the working
directory:

```sh
sudo install -d -o demeter -g demeter /var/lib/demeter
cd /var/lib/demeter
sudo -u demeter /opt/demeter/demeter server --address 0.0.0.0:8249
```

The timelapse path is hard-coded to `/srv/demeter/timelapse`. Create it and
make it writable by the server user:

```sh
sudo install -d -o demeter -g demeter /srv/demeter/timelapse
```

If the directory is missing but its parent is writable, the server creates
per-camera subdirectories automatically. It does not create a persistent
volume for you.

## Run in Docker

The repository ships a development stack under `dev/`. It is intended for
testing the server, the web panel, and the Home Assistant integration, not as
a hardened production deployment.

Build the binary first, then start the real-server profile:

```sh
task build
./dev/serve.sh
```

The script starts the Compose `live` profile with:

- The binary bind-mounted from `DEMETER_BINARY` (default `../demeter`).
- SQLite stored on the `demeter-data` Docker volume mounted at `/data`.
- Port 8249 published to the host.
- Home Assistant available at port 8123.

| Variable | Default | Purpose |
| --- | --- | --- |
| `DEMETER_BINARY` | `../demeter` | Binary mounted into the server container. |
| `DEMETER_SERVER_PORT` | `8249` | Host port for the Demeter server. |
| `HA_PORT` | `8123` | Host port for Home Assistant. |

The development stack does **not** mount `/srv/demeter/timelapse`. Images are
lost when the container is recreated. Add a volume or bind mount for
production use.

Stop it with:

```sh
./dev/down.sh
```

## Security before you start

Demeter has no authentication. Every route, including writes and deletes, is
open to anyone who can reach the port. The sensor WebSocket is also
unauthenticated.

For a first local run, binding to `localhost` is fine:

```sh
./demeter server --address 127.0.0.1:8249
```

For a LAN install, bind to a private interface and restrict access with a
firewall or VLAN. Never port-forward Demeter directly from the internet. If
you need remote access, use a VPN or an authenticating TLS reverse proxy. See
[Operating and Maintaining Demeter](../operating-and-maintaining-demeter#exposing-the-port-safely).

## Verify the server is running

Check the API:

```sh
curl -i http://localhost:8249/api/enclosures
```

A fresh database returns `200 OK` and a JSON `null` rather than `[]`:

```json
null
```

Most list endpoints behave this way because the Go handler marshals a nil
slice. Treat `null` and `[]` as the same empty result in your own clients.

Then open the panel:

```text
http://localhost:8249
```

If the port is unreachable, confirm the process is still running and that the
address includes the interface you are connecting to. See
[Troubleshooting](../troubleshooting).

## Create the first records

The panel opens on an empty dashboard. Create the hierarchy before sending
sensor data:

1. **Enclosures &rarr; New Enclosure**.
2. **Systems &rarr; New System**, choosing the enclosure.
3. **Flows &rarr; New Flow**, choosing the system.
4. **Plants &rarr; New Plant Site**, choosing the flow and coordinates.

The numeric IDs assigned to these records are what sensors use as targets.
Write them down.

## Feed it some data

The binary includes a synthetic sensor client:

```sh
./demeter simulate-sensor --addr localhost:8249
```

It connects to the WebSocket and sends every scalar message type for **target
ID 1**, every 5 seconds by default. It does not send camera frames:

```sh
./demeter simulate-sensor --addr localhost:8249 --interval 10s
```

For readings to appear under real records, the database needs an enclosure,
system, flow, and plant site with ID `1`. The demo data does not create them for
you. The server does not enforce foreign keys, so readings with missing targets
are still stored as orphan rows; they just do not appear in any entity's
charts. The simulator is also available as `DEMETER_ADDR=localhost:8249
./demeter simulate-sensor`.

In the Docker stack:

```sh
docker compose -f dev/docker-compose.yml exec demeter \
  /usr/local/bin/demeter simulate-sensor --addr demeter:8249
```

Open an enclosure, system, or plant page and refresh after a batch. Enclosure
air charts use the server clock and a selectable time window (default: last 24
hours). System charts plot up to 20 recent raw readings, and plant pages plot
up to 100 raw lux readings.

## Next steps

- [Using the Web Control Panel](../using-the-web-control-panel)
- [Enclosures](../enclosures)
- [ESP32 Hardware](../esp32-hardware)
- [Building and Flashing the ESP32 Firmware](../building-and-flashing-the-esp32-firmware)
- [Connecting to Home Assistant](../connecting-to-home-assistant)
- [Operating and Maintaining Demeter](../operating-and-maintaining-demeter)
