---
title: "Operating and Maintaining Demeter"
summary: "Run Demeter as a service, back up and restore SQLite and timelapse data, upgrade safely, and build the documentation."
date: 2023-09-07T16:13:18+02:00
lastmod: 2026-11-03T00:00:00+00:00
draft: false
weight: 16
toc: true
---

This page covers the boring but important parts: running the server
reliably, protecting its data, upgrading it, and building these docs.

<figure>
  <img src="/images/sample.jpg" alt="Service status and backups" width="1200" height="675" loading="lazy">
  <figcaption>A production install runs Demeter as a service with a persistent data directory. Replace this placeholder with a screenshot.</figcaption>
</figure>

## Runtime layout

Demeter has only two pieces of persistent state:

| State | Default location | Notes |
| --- | --- | --- |
| SQLite database | `./demeter.sqlite` | Relative to the process working directory. |
| Timelapse images | `/srv/demeter/timelapse/` | Absolute path, hard-coded in the server. |

Everything else — the REST API, WebSocket endpoint, web UI, and SQL schema —
is inside the compiled binary.

There is no configuration file. The server address is the only runtime
setting:

```sh
./demeter server --address 0.0.0.0:8249
```

The database path is not configurable. Choose the process working directory
to choose where `demeter.sqlite` lives.

## Run under systemd

A minimal unit assumes:

- The binary is at `/opt/demeter/demeter`.
- Data lives in `/var/lib/demeter`.
- Logs go to the journal.

```ini
# /etc/systemd/system/demeter.service
[Unit]
Description=Demeter hydroponics server
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=demeter
Group=demeter
WorkingDirectory=/var/lib/demeter
ExecStart=/opt/demeter/demeter server --address 0.0.0.0:8249
Restart=on-failure
RestartSec=5s

# The timelapse root is hard-coded, so make it writable for this user.
ReadWritePaths=/var/lib/demeter /srv/demeter/timelapse

[Install]
WantedBy=multi-user.target
```

Prepare the directories and service account:

```sh
sudo useradd --system --home /var/lib/demeter --shell /usr/sbin/nologin demeter
sudo install -d -o demeter -g demeter /var/lib/demeter /srv/demeter/timelapse
sudo install -m 0755 demeter /opt/demeter/demeter
sudo systemctl daemon-reload
sudo systemctl enable --now demeter
sudo systemctl status demeter
```

Follow logs with:

```sh
journalctl -u demeter -f
```

The server logs to stderr, which systemd captures. There is no log rotation
inside Demeter; journald handles it.

### Exposing the port safely

`0.0.0.0` makes the server reachable on every interface. Because Demeter has
no authentication, prefer one of:

- Bind to a specific LAN address: `--address 192.168.1.20:8249`.
- Keep it on a trusted VLAN.
- Put a reverse proxy in front of it for TLS and authentication.
- Use a VPN for remote access.

If you use a reverse proxy, proxy both the REST API and the sensor WebSocket,
including the `Upgrade` and `Connection` headers.

## Back up and restore

### Database

Use SQLite's online backup command while the server is running:

```sh
sqlite3 /var/lib/demeter/demeter.sqlite \
  ".backup '/var/backups/demeter-$(date +%F).sqlite'"
```

For a fully cold backup, stop the service first and copy the file:

```sh
sudo systemctl stop demeter
cp /var/lib/demeter/demeter.sqlite /var/backups/demeter-$(date +%F).sqlite
sudo systemctl start demeter
```

The database contains topology, readings, notes, and timestamps. It does not
contain timelapse images.

### Timelapse images

```sh
tar -C /srv/demeter -czf /var/backups/demeter-timelapse-$(date +%F).tar.gz timelapse
```

Images are large and grow without automatic pruning. Back them up separately
from the database and apply a retention policy that matches your needs.

### Restore

```sh
sudo systemctl stop demeter
sudo install -o demeter -g demeter -m 0640 \
  /var/backups/demeter-2026-11-03.sqlite \
  /var/lib/demeter/demeter.sqlite
sudo systemctl start demeter
```

Test a restore periodically on a non-production host. A backup you have never
restored is only a hope.

## Upgrade

1. **Back up** the database and timelapse directory.
2. Build or download the new binary.
3. Stop the service.
4. Replace the binary.
5. Start the service.
6. Check the log and API.

```sh
sudo systemctl stop demeter
sudo install -m 0755 demeter /opt/demeter/demeter
sudo systemctl start demeter
curl -fsS http://localhost:8249/api/enclosures >/dev/null && echo OK
```

On startup the server tries to execute the embedded `schema.sql`, but it
ignores the result. On a fresh database this creates the tables. On an existing
database the script normally stops at the first table that already exists, and
the server does not perform schema migrations or reliably add missing tables.
If a release changes the schema, follow that release's migration notes, which
may require SQLite commands or a fresh database. This is another reason to keep
backups.

The web UI is embedded in the binary. Replacing the binary updates the UI; no
separate web deployment is needed.

## Disk usage and retention

There is no built-in retention for readings or timelapse images.

- **Readings** accumulate indefinitely in SQLite.
- **Images** accumulate indefinitely under `/srv/demeter/timelapse`.
- **Logs** are handled by journald or your container runtime.

Monitor disk space:

```sh
df -h /var/lib/demeter /srv/demeter/timelapse
du -sh /var/lib/demeter/demeter.sqlite /srv/demeter/timelapse
```

Prune old timelapse files with a cron job or systemd timer:

```sh
find /srv/demeter/timelapse -type f -name '*.jpeg' -mtime +30 -delete
```

If you need to prune old readings, do it in a transaction and back up first:

```sh
sqlite3 /var/lib/demeter/demeter.sqlite <<'SQL'
BEGIN;
DELETE FROM system_ph WHERE timestamp < datetime('now', '-365 days');
DELETE FROM system_ec WHERE timestamp < datetime('now', '-365 days');
DELETE FROM system_water_temperature WHERE timestamp < datetime('now', '-365 days');
DELETE FROM enclosure_air_temperature WHERE timestamp < datetime('now', '-365 days');
DELETE FROM enclosure_air_humidity WHERE timestamp < datetime('now', '-365 days');
DELETE FROM plant_site_lux WHERE timestamp < datetime('now', '-365 days');
COMMIT;
VACUUM;
SQL
```

Confirm table and column names against your schema before running deletions.
`VACUUM` temporarily needs free space roughly equal to the database size.

## Health checks

A simple HTTP check:

```sh
curl -fsS http://localhost:8249/api/enclosures >/dev/null
```

A deeper check can query a known system and verify that recent readings exist:

```sh
curl -fsS 'http://localhost:8249/api/systems/1/ph?limit=1'
```

An empty result (`null` or `[]`) is still a successful HTTP response. For
monitoring, alert on:

- Process not running.
- HTTP check failing.
- No new raw reading for longer than twice the sensor interval.
- Disk usage above your threshold.

## Using the simulator in production-like testing

The binary includes a sensor simulator. It connects to the WebSocket and sends
synthetic readings for **target ID 1** across every message type:

```sh
./demeter simulate-sensor --addr localhost:8249
./demeter simulate-sensor --addr localhost:8249 --interval 10s
```

It respects `DEMETER_ADDR` as an alternative to `--addr`, and it sends an
Initialize frame first followed by every scalar message type. It does not send
camera frames. The simulator needs an enclosure, system, flow, and
plant site with ID `1` for readings to appear under real records. Because
foreign keys are not enforced, readings with missing targets are accepted as
orphan rows and simply do not show up in the UI. Use it against a disposable
database, not production.

## Docker

The repository's `dev/docker-compose.yml` is a **development and test harness**,
not a hardened deployment. Its `live` profile runs the real binary with:

- A bind-mounted binary (`DEMETER_BINARY`).
- A named volume `demeter-data` mounted at `/data` for SQLite.
- Port `8249` published to the host.

It does **not** mount `/srv/demeter/timelapse`. Timelapse images inside that
container are lost when it is recreated. For a containerised production
deployment, add a volume for the timelapse root and run the container as a
non-root user with write access to both paths.

See [Installing the Server](../installing-the-server#run-in-docker).

## Building the documentation

The docs site is a Hugo site under `docs/` using Doks/Thulite.

Install dependencies and build:

```sh
cd docs
npm install
npm run build
```

The output is written to `docs/public`. To preview with live reload:

```sh
cd docs
npm run dev
```

The dev server runs at <http://localhost:1313>. The OpenAPI reference is
generated from `docs/static/openapi.yaml`:

```sh
cd docs
task openapi
```

From the repository root, the included tasks are:

```sh
task docs:dev
task site
```

`task site` runs the docs dev server directly. `task docs:dev` regenerates the
OpenAPI HTML first.

### Updating API docs

When you add or change an endpoint:

1. Update `docs/static/openapi.yaml`.
2. Run `task openapi` in `docs/`.
3. Update the endpoint index in [REST API](../rest-api).
4. Rebuild and check the page.

## Security checklist

- [ ] Demeter is not reachable from the public internet.
- [ ] The sensor WebSocket and REST API are on a trusted network.
- [ ] A reverse proxy provides TLS and authentication if remote access is
      needed.
- [ ] The database and timelapse directory are readable only by the service
      account.
- [ ] Backups are encrypted if they leave the host.
- [ ] WiFi credentials and `esp-sensor/sdkconfig` are not committed to a public
      repository.
- [ ] The server clock is synchronised with NTP.

## Related guides

- [Installing the Server](../installing-the-server)
- [Troubleshooting](../troubleshooting)
- [REST API](../rest-api)
- [Photos and Timelapse](../photos-and-timelapse)
