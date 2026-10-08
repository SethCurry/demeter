# demeter

## WARNING

Demeter is under heavy construction.  Do not deploy it
as is.

The goal is to get an initial, usable release out within the next month.

Stay tuned for updates

## Goal

Make it easy to monitor hydroponics, get timelapse photos, and perform automated maintenance.

## Design

### Monitoring

- pH
- EC
- O2 %
- Water level
- Water flow rate
- Light intensity
- Water temperature
- Air temperature
- Air humidity

Derived:

- Total volume of reservoir (from water depth + dimensions)

### Timelapse Photos

- Of whole system
- Of roots

### Automated Maintenance

- Top up water
- Do routine drains of water

## Schema

Enclosures are the space the thing is in, they should all share air qualities
Systems are defined by sharing a reservoir
Loops share a single flow of water
Plant site is a single location within that loop
