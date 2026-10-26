---
title: "Connecting to Home Assistant"
summary: "Install the Demeter custom integration into Home Assistant and expose every enclosure, system, and plant site as sensor entities you can dashboard and automate against."
date: 2023-09-07T16:13:18+02:00
lastmod: 2023-09-07T16:13:18+02:00
draft: false
weight: 3
toc: true
---

Demeter integrates with [Home Assistant](https://www.home-assistant.io)
through a small custom integration that ships in the repository under
[`python/homeassistant/`](https://github.com/SethCurry/demeter/tree/main/python/homeassistant).
It is a **hub** integration of the `local_poll` IoT class: it connects to a
running Demeter server over HTTP, discovers your topology on every poll, and
exposes each enclosure, system, and plant site as a Home Assistant device with
one sensor entity per metric. The Demeter API has no authentication, so all
you configure is the server's base URL.

This guide walks through installing the integration, adding it through the
config flow, what you get in return, and how to keep it healthy.

## Prerequisites

Before you start you need:

- A **Demeter server** that is up and reachable over HTTP &mdash; see
  [Installing the Server](../02-installing-the-server). The integration
  talks to the same REST API the web panel uses, so if you can open
  `http://<your-demeter-host>:8249` in a browser, Home Assistant can reach it
  too.
- A **Home Assistant** instance you administer, running a recent release. The
  integration is tested against the official `stable` image.
- Access to Home Assistant's **configuration directory** (the folder that
  contains `configuration.yaml`), so you can install a custom component.

> **Networking note:** if your Home Assistant runs in a container and Demeter
> runs on the host, point it at `http://host.docker.internal:8249` rather than
> `localhost`. If they are both containerised on the same Compose network, use
> the service name, e.g. `http://demeter:8249`.

## How it works

When you add the integration Home Assistant creates a config entry holding
the Demeter base URL. On setup &mdash; and once a minute thereafter &mdash;
the integration's data-update coordinator:

1. Discovers the full topology by calling
   `/api/enclosures`, `/api/systems`, `/api/flows`, and `/api/plant-sites`.
2. For every device that exposes a metric, fetches the latest reading from
   that device's reading endpoint (for example
   `/api/systems/{id}/ph` or `/api/enclosures/{id}/air-temperature`).
3. Publishes the values to one sensor entity per metric, tagged with a
   `reading_timestamp` attribute carrying the time Demeter recorded the
   reading.

The poll interval is fixed at **60 seconds**. The first refresh discovers the
topology; if Demeter is unreachable at that point the config entry retries
with `ConfigEntryNotReady` rather than failing outright. Devices you add in
Demeter after setup are picked up automatically on the next poll, with no
restart required.

## Install the integration

The integration is a custom component, so you install it by copying its
source into Home Assistant's `custom_components/` directory.

```sh
git clone https://github.com/SethCurry/demeter.git
cd demeter
cp -r python/homeassistant/ ~/.homeassistant/custom_components/demeter/
```

Adjust the destination to match wherever your Home Assistant configuration
lives (`/config/custom_components/demeter/` is common in containerised
installs). The result should be a single directory:

```text
custom_components/
└── demeter/
    ├── __init__.py
    ├── api.py
    ├── config_flow.py
    ├── const.py
    ├── coordinator.py
    ├── manifest.json
    ├── sensor.py
    ├── strings.json
    └── translations/
```

**Restart Home Assistant** so it picks up the new custom component. Until you
restart, "Demeter" will not appear in the Add Integration list.

> The integration has no Python dependencies &mdash; `manifest.json` lists an
> empty `requirements` array &mdash; so there is nothing to install from PyPI.

## Add the integration

Once Home Assistant has restarted, add Demeter through the UI:

1. Open **Settings &rarr; Devices & Services**.
2. Click **Add Integration** (the bottom-right button) and search for
   **Demeter**.
3. The config flow shows a single field, **Host**. Enter the base URL of your
   Demeter server, for example `http://demeter.local:8249`.

   A scheme is optional: if you omit it, `http://` is added for you, so
   `demeter.local:8249` becomes `http://demeter.local:8249`. Trailing slashes
   are stripped.

4. Submit the form. Home Assistant tries to reach `/api/enclosures` to verify
   the connection. If it succeeds, the entry is created; if Demeter is
   unreachable you get a **Failed to connect** error and can try again.

Each Demeter server is keyed by its base URL, so you can only add a given host
once. To monitor more than one Demeter server, repeat the flow for each one.

## What you get

Every Demeter device that exposes metrics becomes one Home Assistant device,
named after the Demeter device and attributed to manufacturer **Demeter**.
Plant site devices are named `<flow> Site (x, y, z)`, matching how they appear
in the web panel. Each device's page links back to the Demeter server via the
configured URL.

| Demeter device | Sensor entities                                   | Unit     |
| -------------- | -------------------------------------------------- | -------- |
| Enclosure      | Air Temperature, Air Humidity                      | &deg;C, % |
| System         | pH, EC, Water Temperature                          | -, &micro;S/cm, &deg;C |
| Plant site     | Illuminance                                        | lx       |

A few details worth knowing:

- All sensors use the `measurement` state class, so they record history you
  can graph in Home Assistant's energy/sensor dashboards.
- Temperature, humidity, and illuminance sensors carry the standard Home
  Assistant device classes, so they render with the right units and icons
  automatically. pH and EC use explicit icons (`mdi:flask` and `mdi:omega`).
- Each entity exposes a **`reading_timestamp`** extra state attribute holding
  the ISO-formatted time Demeter recorded the value &mdash; useful for
  telling a stale reading from a fresh one when the sensor has been offline.
- EC values are stored without a unit annotation in Demeter. The integration
  assumes **&micro;S/cm**, the common unit for hobby EC probes. If Demeter
  settles on a different convention, adjust `DEVICE_METRICS` in
  `sensor.py`.

## Use the data

Once the entities exist they behave like any other Home Assistant sensor. A
few common patterns:

- **Dashboard cards:** add a glance card or a history card for
  `sensor.<system>_ph`, `sensor.<enclosure>_air_temperature`, etc.
- **Automations:** trigger on thresholds, e.g. notify when pH drifts outside
  a band or when air temperature exceeds a setpoint.
- **Long-term statistics:** because the state class is `measurement`,
  temperature, humidity, and illuminance are eligible for long-term
  statistics. pH and EC are unitless/numeric measurements and are recorded as
  plain history.

## Troubleshooting

**Failed to connect during setup.** Confirm the Demeter server is running and
that Home Assistant can reach it from inside its own network context. Test
with the same URL Home Assistant would use:

```sh
curl http://demeter.local:8249/api/enclosures
```

The most common cause is a container networking mismatch &mdash; see the
networking note in [Prerequisites](#prerequisites).

**Sensors show as unknown.** The integration reports `unknown` for a metric
when Demeter has no reading for it yet (for example a freshly created system
with no pH probe reporting). Check the Demeter web panel for that device; the
sensor will populate on the next poll once a reading arrives. A sensor also
goes `unknown` if a single reading endpoint fails mid-poll &mdash; the
coordinator logs a warning and leaves that one entity stale rather than
failing the whole update.

**Newly added Demeter devices do not appear.** They are picked up on the next
60-second poll, so wait at most a minute and refresh the integration's device
list. You do not need to reload the config entry.

**Renaming a device in Demeter does not rename it in Home Assistant.** The
Home Assistant device name is fixed when the entity is first created. To pick
up a new name, remove and re-add the integration, or rename the device
directly in Home Assistant's device registry.

**Reload after editing the source.** If you change files under
`custom_components/demeter/`, restart Home Assistant (or reload the
integration via **Devices & Services &rarr; Demeter &rarr; Reload**) for the
changes to take effect.

## Test it first (optional)

The repository ships a Docker-based test harness under
[`dev/`](https://github.com/SethCurry/demeter/tree/main/dev) that runs Home
Assistant from the official `stable` image with this integration mounted as a
custom component, alongside a mock Demeter API with deterministic readings:

```sh
./dev/test.sh
```

The smoke test completes onboarding, installs the integration through the real
config flow, and verifies the resulting sensor values, units, and reading
timestamps. Use `./dev/up.sh` instead to start the stack for manual UI
exploration. Both scripts mount `python/homeassistant/` by default and accept
the `DEMETER_HA_INTEGRATION` environment variable to override the path.

## Remove the integration

To uninstall:

1. Open **Settings &rarr; Devices & Services &rarr; Demeter** and click
   **Delete**. This removes the config entry and all of its entities.
2. Delete the `custom_components/demeter/` directory.
3. Restart Home Assistant.