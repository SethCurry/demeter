-- name: GetEnclosure :one
SELECT * FROM enclosure WHERE id=? LIMIT 1;

-- name: ListEnclosures :many
SELECT * FROM enclosure;

-- name: CreateEnclosure :one
INSERT INTO enclosure (name)
VALUES (?)
RETURNING *;

-- name: UpdateEnclosure :one
UPDATE enclosure
SET name = ?
WHERE id = ?
RETURNING *;

-- name: DeleteEnclosure :exec
DELETE FROM enclosure WHERE id = ?;

-- name: GetSystem :one
SELECT * FROM system WHERE id=? LIMIT 1;

-- name: ListSystems :many
SELECT * FROM system;

-- name: ListSystemsByEnclosure :many
SELECT * FROM system WHERE enclosure_id = ?;

-- name: CreateSystem :one
INSERT INTO system (name, enclosure_id)
VALUES (?, ?)
RETURNING *;

-- name: UpdateSystem :one
UPDATE system
SET name = ?, enclosure_id = ?
WHERE id = ?
RETURNING *;

-- name: DeleteSystem :exec
DELETE FROM system WHERE id = ?;

-- name: GetFlow :one
SELECT * FROM flow WHERE id=? LIMIT 1;

-- name: ListFlows :many
SELECT * FROM flow;

-- name: ListFlowsBySystem :many
SELECT * FROM flow WHERE system_id = ?;

-- name: ListChildFlows :many
SELECT * FROM flow WHERE parent_flow_id = ?;

-- name: CreateFlow :one
INSERT INTO flow (name, system_id, parent_flow_id)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateFlow :one
UPDATE flow
SET name = ?, system_id = ?, parent_flow_id = ?
WHERE id = ?
RETURNING *;

-- name: DeleteFlow :exec
DELETE FROM flow WHERE id = ?;

-- name: GetPlantSite :one
SELECT * FROM plant_site WHERE id=? LIMIT 1;

-- name: ListPlantSites :many
SELECT * FROM plant_site;

-- name: ListPlantSitesByFlow :many
SELECT * FROM plant_site WHERE flow_id = ?;

-- name: CreatePlantSite :one
INSERT INTO plant_site (flow_id, x, y, z)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdatePlantSite :one
UPDATE plant_site
SET flow_id = ?, x = ?, y = ?, z = ?
WHERE id = ?
RETURNING *;

-- name: DeletePlantSite :exec
DELETE FROM plant_site WHERE id = ?;

-- name: GetPlant :one
SELECT * FROM plant WHERE id=? LIMIT 1;

-- name: ListPlants :many
SELECT * FROM plant;

-- name: ListPlantsBySite :many
SELECT * FROM plant WHERE plant_site_id = ?;

-- name: CreatePlant :one
INSERT INTO plant (plant_site_id, planted_on)
VALUES (?, ?)
RETURNING *;

-- name: UpdatePlant :one
UPDATE plant
SET plant_site_id = ?, planted_on = ?
WHERE id = ?
RETURNING *;

-- name: DeletePlant :exec
DELETE FROM plant WHERE id = ?;

-- name: GetEnclosureNote :one
SELECT * FROM enclosure_note WHERE id=? LIMIT 1;

-- name: ListEnclosureNotes :many
SELECT * FROM enclosure_note;

-- name: CreateEnclosureNote :one
INSERT INTO enclosure_note (timestamp, content)
VALUES (?, ?)
RETURNING *;

-- name: UpdateEnclosureNote :one
UPDATE enclosure_note
SET timestamp = ?, content = ?
WHERE id = ?
RETURNING *;

-- name: DeleteEnclosureNote :exec
DELETE FROM enclosure_note WHERE id = ?;

-- name: GetSystemNote :one
SELECT * FROM system_note WHERE id=? LIMIT 1;

-- name: ListSystemNotes :many
SELECT * FROM system_note;

-- name: CreateSystemNote :one
INSERT INTO system_note (timestamp, content)
VALUES (?, ?)
RETURNING *;

-- name: UpdateSystemNote :one
UPDATE system_note
SET timestamp = ?, content = ?
WHERE id = ?
RETURNING *;

-- name: DeleteSystemNote :exec
DELETE FROM system_note WHERE id = ?;

-- name: GetFlowNote :one
SELECT * FROM flow_note WHERE id=? LIMIT 1;

-- name: ListFlowNotes :many
SELECT * FROM flow_note;

-- name: CreateFlowNote :one
INSERT INTO flow_note (timestamp, content)
VALUES (?, ?)
RETURNING *;

-- name: UpdateFlowNote :one
UPDATE flow_note
SET timestamp = ?, content = ?
WHERE id = ?
RETURNING *;

-- name: DeleteFlowNote :exec
DELETE FROM flow_note WHERE id = ?;

-- name: GetPlantSiteNote :one
SELECT * FROM plant_site_note WHERE id=? LIMIT 1;

-- name: ListPlantSiteNotes :many
SELECT * FROM plant_site_note;

-- name: ListPlantSiteNotesBySite :many
SELECT * FROM plant_site_note WHERE plant_site_id = ?;

-- name: CreatePlantSiteNote :one
INSERT INTO plant_site_note (plant_site_id, timestamp, content)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdatePlantSiteNote :one
UPDATE plant_site_note
SET plant_site_id = ?, timestamp = ?, content = ?
WHERE id = ?
RETURNING *;

-- name: DeletePlantSiteNote :exec
DELETE FROM plant_site_note WHERE id = ?;

-- name: GetPlantNote :one
SELECT * FROM plant_note WHERE id=? LIMIT 1;

-- name: ListPlantNotes :many
SELECT * FROM plant_note;

-- name: ListPlantNotesByPlant :many
SELECT * FROM plant_note WHERE plan_id = ?;

-- name: CreatePlantNote :one
INSERT INTO plant_note (plan_id, timestamp, content)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdatePlantNote :one
UPDATE plant_note
SET plan_id = ?, timestamp = ?, content = ?
WHERE id = ?
RETURNING *;

-- name: DeletePlantNote :exec
DELETE FROM plant_note WHERE id = ?;

-- name: ListEnclosureAirTemperatures :many
SELECT * FROM enclosure_air_temperature
WHERE enclosure_id = ?
ORDER BY timestamp DESC
LIMIT ?;

-- name: CreateEnclosureAirTemperature :exec
INSERT INTO enclosure_air_temperature (enclosure_id, temperature_c)
VALUES (?, ?);

-- name: ListEnclosureAirHumidity :many
SELECT * FROM enclosure_air_humidity
WHERE enclosure_id = ?
ORDER BY timestamp DESC
LIMIT ?;

-- name: CreateEnclosureAirHumidity :exec
INSERT INTO enclosure_air_humidity (enclosure_id, humidity_rh)
VALUES (?, ?);

-- name: CreateSystemPH :exec
INSERT INTO system_ph (system_id, ph)
VALUES (?, ?);

-- name: CreateSystemEC :exec
INSERT INTO system_ec (system_id, ec)
VALUES (?, ?);

-- name: CreateSystemOxygen :exec
INSERT INTO system_oxygen (system_id, oxygen)
VALUES (?, ?);

-- name: CreateSystemWaterTemperature :exec
INSERT INTO system_water_temperature (system_id, water_temperature_c)
VALUES (?, ?);

-- name: CreateSystemWaterProximity :exec
INSERT INTO system_water_proximity (system_id, proximity_m)
VALUES (?, ?);

-- name: CreateFlowRate :exec
INSERT INTO flow_rate (flow_id, rate_mps)
VALUES (?, ?);

-- name: CreatePlantSiteLux :exec
INSERT INTO plant_site_lux (plant_site_id, lux)
VALUES (?, ?);
