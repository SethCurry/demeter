# Demeter Web

Web control panel for the Demeter hydroponics system. Built with React,
TypeScript, [Ant Design](https://ant.design), and React Router.

## Running

From this directory:

```bash
npm install
npm start      # dev server at http://localhost:3000
npm run build  # production build into build/
npm test       # run tests
```

## Pages

The sidebar mirrors the Demeter domain model (Enclosure → System → Flow → Plant Site):

- **Dashboard** — live overview: counts, average pH/EC/light, and per-system status.
- **Enclosures** — spaces that share air qualities (air temp, humidity).
- **Systems** — defined by sharing a reservoir (pH, EC, O₂, water temp, level).
- **Flows** — loops sharing a single flow of water (rate, parent flow, site count).
- **Plants** — a locus within a loop (position, lux, species, planted date).
- **Photos** — timelapse capture of whole systems and roots.
- **Maintenance** — automated top-ups and routine drains.
- **Settings** — backend connection and capture/retention preferences.

## Notes

- Sensor readings are currently mocked in `src/data/mock.ts`. Wire these up to the
  backend's WebSocket (`/api/websocket/sensors`) or future REST endpoints.
- The theme primary color is green (`#2e7d32`); adjust in `src/index.tsx`.