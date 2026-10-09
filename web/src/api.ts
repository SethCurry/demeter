// Typed client for the demeter Go HTTP API (see internal/httpd).
//
// The backend exposes sqlc-generated model structs directly via gin. Those
// structs carry no JSON tags, so response fields are PascalCase. Nullable
// columns use database/sql Null* types, which encoding/json serializes as
// {"Int64": n, "Valid": bool} / {"Time": "...", "Valid": bool} /
// {"String": "...", "Valid": bool} objects. The interfaces and helpers below
// mirror that shape exactly.

const API_BASE = (process.env.REACT_APP_API_BASE as string | undefined) ?? "";

export interface NullInt64 {
  Int64: number;
  Valid: boolean;
}

export interface NullTime {
  Time: string;
  Valid: boolean;
}

export interface NullString {
  String: string;
  Valid: boolean;
}

export interface Enclosure {
  ID: number;
  Name: string;
}

export interface System {
  ID: number;
  Name: string;
  EnclosureID: number;
}

export interface Flow {
  ID: number;
  Name: string;
  SystemID: number;
  ParentFlowID: NullInt64;
}

export interface PlantSite {
  ID: number;
  FlowID: number;
  X: number;
  Y: number;
  Z: number;
}

export interface Plant {
  ID: number;
  PlantSiteID: number;
  PlantedOn: NullTime;
}

export interface PlantWithFlow extends Plant {
  FlowName: string;
  FlowID: number;
  PlantSiteID: number;
  X: number;
  Y: number;
  Z: number;
}

// "Simple" notes (enclosure/system/flow) share the same shape.
export interface SimpleNote {
  ID: number;
  Timestamp: NullTime;
  Content: NullString;
}

export interface PlantSiteNote {
  ID: number;
  PlantSiteID: number;
  Timestamp: NullTime;
  Content: NullString;
}

export interface PlantNote {
  ID: number;
  PlanID: NullInt64;
  Timestamp: NullTime;
  Content: NullString;
}

// Sensor readings. Timestamps use the Nullable time shape emitted by sqlc.
export interface EnclosureAirTemperature {
  ID: number;
  Timestamp: NullTime;
  EnclosureID: number;
  TemperatureC: number;
}

export interface EnclosureAirHumidity {
  ID: number;
  Timestamp: NullTime;
  EnclosureID: number;
  HumidityRh: number;
}

async function getJson<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { Accept: "application/json" },
  });
  if (!res.ok)
    throw new Error(`GET ${path} failed: ${res.status} ${res.statusText}`);
  return (await res.json()) as T;
}

async function sendJson<T>(
  method: string,
  path: string,
  body?: unknown,
): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  if (res.status === 204) return undefined as T;
  if (!res.ok) {
    let detail = `${res.status} ${res.statusText}`;
    try {
      const payload = await res.json();
      if (payload && typeof payload.error === "string") detail = payload.error;
    } catch {
      /* ignore non-JSON error bodies */
    }
    throw new Error(`${method} ${path} failed: ${detail}`);
  }
  return (await res.json()) as T;
}

function withQuery(
  base: string,
  params: Record<string, string | number | undefined>,
): string {
  const qs = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => {
    if (v !== undefined && v !== null && String(v) !== "") qs.set(k, String(v));
  });
  const s = qs.toString();
  return s ? `${base}?${s}` : base;
}

export const api = {
  // Enclosures
  listEnclosures: () => getJson<Enclosure[]>("/api/enclosures"),
  getEnclosure: (id: number) => getJson<Enclosure>(`/api/enclosures/${id}`),
  listEnclosureAirTemperatures: (id: number, limit?: number) =>
    getJson<EnclosureAirTemperature[]>(
      withQuery(`/api/enclosures/${id}/air-temperature`, { limit }),
    ),
  listEnclosureAirHumidity: (id: number, limit?: number) =>
    getJson<EnclosureAirHumidity[]>(
      withQuery(`/api/enclosures/${id}/air-humidity`, { limit }),
    ),
  createEnclosure: (name: string) =>
    sendJson<Enclosure>("POST", "/api/enclosures", { name }),
  updateEnclosure: (id: number, name: string) =>
    sendJson<Enclosure>("PUT", `/api/enclosures/${id}`, { name }),
  deleteEnclosure: (id: number) =>
    sendJson<void>("DELETE", `/api/enclosures/${id}`),
  enclosurePlants: (id: number, limit?: number) =>
    getJson<PlantWithFlow[]>(
      withQuery(`/api/enclosures/${id}/plants`, { limit }),
    ),

  // Systems
  listSystems: (enclosureId?: number) =>
    getJson<System[]>(withQuery("/api/systems", { enclosure_id: enclosureId })),
  getSystem: (id: number) => getJson<System>(`/api/systems/${id}`),
  createSystem: (name: string, enclosureId: number) =>
    sendJson<System>("POST", "/api/systems", {
      name,
      enclosure_id: enclosureId,
    }),
  updateSystem: (id: number, name: string, enclosureId: number) =>
    sendJson<System>("PUT", `/api/systems/${id}`, {
      name,
      enclosure_id: enclosureId,
    }),
  deleteSystem: (id: number) => sendJson<void>("DELETE", `/api/systems/${id}`),

  // Flows
  listFlows: (params?: { systemId?: number; parentFlowId?: number }) =>
    getJson<Flow[]>(
      withQuery("/api/flows", {
        system_id: params?.systemId,
        parent_flow_id: params?.parentFlowId,
      }),
    ),
  getFlow: (id: number) => getJson<Flow>(`/api/flows/${id}`),
  createFlow: (name: string, systemId: number, parentFlowId?: number) =>
    sendJson<Flow>("POST", "/api/flows", {
      name,
      system_id: systemId,
      parent_flow_id: parentFlowId,
    }),
  updateFlow: (
    id: number,
    name: string,
    systemId: number,
    parentFlowId?: number,
  ) =>
    sendJson<Flow>("PUT", `/api/flows/${id}`, {
      name,
      system_id: systemId,
      parent_flow_id: parentFlowId,
    }),
  deleteFlow: (id: number) => sendJson<void>("DELETE", `/api/flows/${id}`),
  flowPlants: (id: number) =>
    getJson<PlantWithFlow[]>(`/api/flows/${id}/plants`),

  // Plant sites
  listPlantSites: (flowId?: number) =>
    getJson<PlantSite[]>(withQuery("/api/plant-sites", { flow_id: flowId })),
  getPlantSite: (id: number) => getJson<PlantSite>(`/api/plant-sites/${id}`),
  createPlantSite: (flowId: number, x: number, y: number, z: number) =>
    sendJson<PlantSite>("POST", "/api/plant-sites", {
      flow_id: flowId,
      x,
      y,
      z,
    }),
  updatePlantSite: (
    id: number,
    flowId: number,
    x: number,
    y: number,
    z: number,
  ) =>
    sendJson<PlantSite>("PUT", `/api/plant-sites/${id}`, {
      flow_id: flowId,
      x,
      y,
      z,
    }),
  deletePlantSite: (id: number) =>
    sendJson<void>("DELETE", `/api/plant-sites/${id}`),

  // Plants
  listPlants: (plantSiteId?: number) =>
    getJson<Plant[]>(withQuery("/api/plants", { plant_site_id: plantSiteId })),
  getPlant: (id: number) => getJson<Plant>(`/api/plants/${id}`),
  createPlant: (plantSiteId: number, plantedOn?: string) =>
    sendJson<Plant>("POST", "/api/plants", {
      plant_site_id: plantSiteId,
      planted_on: plantedOn,
    }),
  updatePlant: (id: number, plantSiteId: number, plantedOn?: string) =>
    sendJson<Plant>("PUT", `/api/plants/${id}`, {
      plant_site_id: plantSiteId,
      planted_on: plantedOn,
    }),
  deletePlant: (id: number) => sendJson<void>("DELETE", `/api/plants/${id}`),

  // Notes
  listEnclosureNotes: () => getJson<SimpleNote[]>("/api/enclosure-notes"),
  createEnclosureNote: (content: string, timestamp?: string) =>
    sendJson<SimpleNote>("POST", "/api/enclosure-notes", {
      content,
      timestamp,
    }),

  listSystemNotes: () => getJson<SimpleNote[]>("/api/system-notes"),
  createSystemNote: (content: string, timestamp?: string) =>
    sendJson<SimpleNote>("POST", "/api/system-notes", { content, timestamp }),

  listFlowNotes: () => getJson<SimpleNote[]>("/api/flow-notes"),
  createFlowNote: (content: string, timestamp?: string) =>
    sendJson<SimpleNote>("POST", "/api/flow-notes", { content, timestamp }),

  listPlantSiteNotes: (plantSiteId?: number) =>
    getJson<PlantSiteNote[]>(
      withQuery("/api/plant-site-notes", { plant_site_id: plantSiteId }),
    ),
  createPlantSiteNote: (
    plantSiteId: number,
    content: string,
    timestamp?: string,
  ) =>
    sendJson<PlantSiteNote>("POST", "/api/plant-site-notes", {
      plant_site_id: plantSiteId,
      content,
      timestamp,
    }),

  listPlantNotes: (planId?: number) =>
    getJson<PlantNote[]>(withQuery("/api/plant-notes", { plan_id: planId })),
  createPlantNote: (
    planId: number | null,
    content: string,
    timestamp?: string,
  ) =>
    sendJson<PlantNote>("POST", "/api/plant-notes", {
      plan_id: planId ?? undefined,
      content,
      timestamp,
    }),
};

// --- Nullable field unwrap helpers -----------------------------------------

export const nullInt = (n: NullInt64 | null | undefined): number | null =>
  n && n.Valid ? n.Int64 : null;

export const nullTime = (t: NullTime | null | undefined): string | null =>
  t && t.Valid ? t.Time : null;

export const nullString = (s: NullString | null | undefined): string | null =>
  s && s.Valid ? s.String : null;
