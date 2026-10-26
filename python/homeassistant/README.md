# Demeter Home Assistant integration

A [Home Assistant](https://www.home-assistant.io) integration for
[demeter](https://github.com/SethCurry/demeter). It connects over HTTP to a
running demeter server, discovers every enclosure, system, and plant site,
and exposes their metrics as sensor entities.

## Install

Copy this directory to `<HA config>/custom_components/demeter/` and restart
Home Assistant:

```sh
cp -r python/homeassistant/ ~/.homeassistant/custom_components/demeter/
```

## Configuration

Add the integration via **Settings → Devices & Services → Add Integration →
Demeter** and enter the base URL of the demeter server, e.g.
`http://demeter.local:8249` (a scheme of `http://` is assumed if omitted).

## What you get

| Demeter device | Sensor entities                                     |
| --------------- | --------------------------------------------------- |
| Enclosure       | Air Temperature (°C), Air Humidity (%)              |
| System          | pH, EC (µS/cm), Water Temperature (°C)               |
| Plant site      | Illuminance (lx)                                    |

Each demeter device becomes one Home Assistant device named after it; plant
site devices are named `<flow> Site (x, y, z)`. Entities are polled once a
minute from the reading endpoints of the demeter API
(`/api/enclosures/{id}/air-temperature`, `/api/systems/{id}/ph`, ...), and
each sensor carries a `reading_timestamp` attribute with the time demeter
recorded the reading. Devices added in demeter after setup are picked up
automatically on the next poll.

## Testing

A Docker-based test harness lives in the repository root at
[`dev/`](../../dev/README.md). It runs Home Assistant from the official
`ghcr.io/home-assistant/home-assistant:stable` image with this integration
mounted as a custom component, plus a mock demeter API with deterministic
readings:

```sh
./dev/test.sh
```

The smoke test completes onboarding, installs the integration through the
real config flow, and verifies the resulting sensor values, units, and
reading timestamps. Use `./dev/up.sh` to start the stack for manual UI
testing instead. Both scripts mount this directory by default and accept
`DEMETER_HA_INTEGRATION` to override the path.

## Removing the integration

Remove the config entry via **Settings → Devices & Services → Demeter →
Delete**, then delete `custom_components/demeter/` and restart Home
Assistant.

Notes:

- demeter's API has no authentication, so only the host is configured.
- EC readings are unitless in demeter; the integration assumes µS/cm.
- Renaming a device in demeter does not yet rename the Home Assistant
  device (it is fixed when the entity is first created).