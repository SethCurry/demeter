# demeter test harness

Shared development/test tooling that lives at the repository root so it can be
used to exercise any part of demeter, not just a single integration.

It currently provides two things:

- a **mock demeter HTTP API** (`mock_demeter.py`) with a fixed topology and
  deterministic readings, for testing demeter clients without a physical
  server or a seeded database;
- a **Home Assistant test stack** (`docker-compose.yml`) that runs the
  official Home Assistant image with the demeter integration mounted as a
  custom component, plus the mock API.

```text
dev/
├── docker-compose.yml   # Home Assistant + mock demeter stack
├── up.sh                # start the stack for manual testing
├── down.sh              # stop the stack
├── test.sh              # start the stack and run smoke_test.py
├── smoke_test.py        # end-to-end Home Assistant integration test
├── mock_demeter.py      # mock demeter HTTP API
├── mock.sh              # run the mock API on the host, no Docker
├── config/              # Home Assistant config (only configuration.yaml is tracked)
└── README.md
```

## Requirements

- Docker with the Compose plugin (for the Home Assistant stack)
- Python 3 on the host (only the standard library is used)

## Mock demeter API

`mock_demeter.py` serves the endpoints used by demeter clients, with the same
response shapes as the sqlc-generated Go structs: PascalCase fields and
`sql.NullTime` timestamps serialised as `{"Time": "...", "Valid": true}`.

| Endpoint                                | Value    |
| --------------------------------------- | -------- |
| `/api/enclosures`                       | Test Enclosure (id 1) |
| `/api/systems`                          | Test System (id 1, enclosure 1) |
| `/api/flows`                            | Test Flow (id 1, system 1) |
| `/api/plant-sites`                      | Site (id 1, flow 1, x/y/z 1/2/3) |
| `/api/enclosures/1/air-temperature`     | 22.5 °C  |
| `/api/enclosures/1/air-humidity`        | 55.0 %   |
| `/api/systems/1/ph`                     | 6.3      |
| `/api/systems/1/ec`                     | 1.8 µS/cm |
| `/api/systems/1/water-temperature`      | 19.5 °C  |
| `/api/plant-sites/1/lux`                | 12000 lx |

Run it directly on the host:

```sh
./dev/mock.sh              # http://localhost:8249
PORT=9000 ./dev/mock.sh    # different port
```

Or as a container without the rest of the stack:

```sh
docker run --rm -p 18249:8249 \
    -v "$PWD/dev/mock_demeter.py:/app/mock_demeter.py:ro" \
    python:3.13-alpine python -u /app/mock_demeter.py
```

## Home Assistant integration test

`./dev/test.sh` runs the end-to-end test for the integration in
`python/homeassistant/`:

1. starts Home Assistant and the mock demeter API;
2. completes Home Assistant onboarding (or logs in on later runs);
3. installs the integration through the same config flow the UI uses;
4. waits for the config entry to load;
5. verifies all six expected sensors, including values, units, and the
   `reading_timestamp` attribute;
6. stops the stack when finished.

```sh
./dev/test.sh
```

Set `KEEP=1` to leave the stack running after the test:

```sh
KEEP=1 ./dev/test.sh
```

The same tasks are available through the Taskfile:

```sh
task ha:test   # ./dev/test.sh
task ha:dev    # ./dev/up.sh
task ha:down   # ./dev/down.sh
```

### Environment variables

| Variable          | Default                   | Purpose                                  |
| ----------------- | ------------------------- | ---------------------------------------- |
| `DEMETER_HA_INTEGRATION` | `../python/homeassistant` | Integration checkout mounted into Home Assistant |
| `HA_PORT`         | `8123`                    | Host port for Home Assistant             |
| `HA_UID`/`HA_GID` | host user/group           | Owner for files Home Assistant writes    |
| `DEMETER_PORT`    | `18249`                   | Host port for the mock demeter API       |
| `HA_URL`          | `http://localhost:8123`   | Home Assistant URL used by the test      |
| `DEMETER_URL`     | `http://demeter-mock:8249`| Host entered in the integration config flow |
| `HA_USERNAME`     | `demeter-test`            | Test account username                    |
| `HA_PASSWORD`     | `demeter-test`            | Test account password                    |
| `KEEP`            | `0`                       | Keep containers after the test           |
| `HA_READY_TIMEOUT`| `180`                     | Seconds to wait for Home Assistant       |
| `HA_SETUP_TIMEOUT`| `60`                      | Seconds to wait for the config entry     |
| `HA_STATE_TIMEOUT`| `90`                      | Seconds to wait for sensor states        |

## Manual testing

Start the stack and leave it running:

```sh
./dev/up.sh
```

Open <http://localhost:8123>. The first launch asks you to create an account;
the smoke test creates `demeter-test` / `demeter-test` if you want to reuse
it. Add the integration via **Settings → Devices & Services → Add
Integration → Demeter** and use:

```text
http://demeter-mock:8249
```

The mock topology contains one enclosure, one system, one flow and one plant
site, producing these entities:

| Entity ID                                      | Value   |
| ---------------------------------------------- | ------- |
| `sensor.test_enclosure_air_temperature`         | 22.5 °C |
| `sensor.test_enclosure_air_humidity`            | 55.0 %  |
| `sensor.test_system_ph`                         | 6.3     |
| `sensor.test_system_ec`                         | 1.8 μS/cm |
| `sensor.test_system_water_temperature`          | 19.5 °C |
| `sensor.test_flow_site_1_2_3_illuminance`       | 12000 lx |

Watch logs with `docker compose -f dev/docker-compose.yml logs -f
homeassistant`, or stop everything with `./dev/down.sh`.

## Testing against a real demeter server

The mock is convenient for integration testing, but you can also point the
integration at a real demeter server running on the host. The Home Assistant
container has a `host.docker.internal` entry, so add the integration with:

```text
http://host.docker.internal:8249
```

If demeter is running elsewhere on your network, use its address directly
instead, as long as it is reachable from the Home Assistant container.

The automated smoke test is written against the mock's fixed values, so it
only works with the mock service.

## Extending the harness

- **Testing another Home Assistant integration layout:** set
  `DEMETER_HA_INTEGRATION` to the directory to mount before running
  `test.sh`, e.g. `DEMETER_HA_INTEGRATION=../some/other/integration`.
- **Testing a non-Home-Assistant client** (the Go server, the web UI,
  scripts): run `./dev/mock.sh` and point the client at
  `http://localhost:8249`. Add new endpoints or a different fixture topology
  to `mock_demeter.py` as needed.
- **Adding a new stack:** put a Compose service in `docker-compose.yml` and a
  matching `*.sh` entry point next to the existing scripts. Keep fixtures and
  deterministic data in this directory so tests stay reproducible.
