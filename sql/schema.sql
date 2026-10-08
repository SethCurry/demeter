CREATE TABLE enclosure (
	id INTEGER PRIMARY KEY,
	name TEXT UNIQUE NOT NULL
);

CREATE TABLE enclosure_note (
    id INTEGER PRIMARY KEY,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    content TEXT
);

CREATE TABLE enclosure_air_temperature (
	id INTEGER PRIMARY KEY,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	enclosure_id INTEGER REFERENCES enclosure(id) NOT NULL,
	temperature_c FLOAT NOT NULL
);

CREATE TABLE enclosure_air_humidity (
	id INTEGER PRIMARY KEY,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	enclosure_id INTEGER REFERENCES enclosure(id) NOT NULL,
	humidity_rh FLOAT NOT NULL
);


CREATE TABLE system (
	id INTEGER PRIMARY KEY,
	name TEXT UNIQUE NOT NULL,
	enclosure_id INTEGER REFERENCES enclosure(id) NOT NULL
);

CREATE TABLE system_note (
    id INTEGER PRIMARY KEY,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    content TEXT
);

CREATE TABLE system_ph (
	id INTEGER PRIMARY KEY,
	system_id INTEGER REFERENCES system(id) NOT NULL,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	ph FLOAT NOT NULL
);

CREATE TABLE system_ec (
	id INTEGER PRIMARY KEY,
	system_id INTEGER REFERENCES system(id) NOT NULL,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	ec FLOAT NOT NULL
);

CREATE TABLE system_oxygen (
	id INTEGER PRIMARY KEY,
	system_id INTEGER REFERENCES system(id) NOT NULL,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	oxygen FLOAT NOT NULL
);

CREATE TABLE system_water_temperature (
	id INTEGER PRIMARY KEY,
	system_id INTEGER REFERENCES system(id) NOT NULL,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	water_temperature_c FLOAT NOT NULL
);

CREATE TABLE system_water_proximity (
	id INTEGER PRIMARY KEY,
	system_id INTEGER REFERENCES system(id) NOT NULL,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	proximity_m FLOAT NOT NULL
);

CREATE TABLE flow (
	id INTEGER PRIMARY KEY,
	name TEXT UNIQUE NOT NULL,
	system_id INTEGER REFERENCES system(id) NOT NULL,
	parent_flow_id INTEGER REFERENCES flow(id)
);

CREATE TABLE flow_note (
    id INTEGER PRIMARY KEY,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    content TEXT
);

CREATE TABLE flow_rate (
	id INTEGER PRIMARY KEY,
	flow_id INTEGER REFERENCES flow(id) NOT NULL,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	rate_mps FLOAT NOT NULL
);

CREATE TABLE plant_site (
	id INTEGER PRIMARY KEY,
	flow_id INTEGER REFERENCES flow(id) NOT NULL,

	-- left to right
	x INTEGER NOT NULL,

	-- bottom to top
	y INTEGER NOT NULL,

	-- back to front
	z INTEGER NOT NULL
);

CREATE TABLE plant_site_note (
    id INTEGER PRIMARY KEY,
    plant_site_id INTEGER REFERENCES plant_site(id) NOT NULL,
    timestamp DATETIME NOT NULL,
    content TEXT
);

CREATE TABLE plant_site_lux (
	id INTEGER PRIMARY KEY,
	plant_site_id INTEGER REFERENCES plant_site(id) NOT NULL,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	lux INTEGER NOT NULL
);

CREATE TABLE plant_genus (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE plant_species (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    plant_genus_id INTEGER REFERENCES plant_genus(id)
);

CREATE TABLE plant (
    id INTEGER PRIMARY KEY,
    plant_site_id INTEGER REFERENCES plant_site(id) NOT NULL,
    planted_on DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE plant_note (
    id INTEGER PRIMARY KEY,
    plan_id INTEGER REFERENCES plant(id),
    timestamp DATETIME NOT NULL,
    content TEXT
);