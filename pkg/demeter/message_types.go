package demeter

// MessageTypeID identifies the kind of payload carried by a message so that
// receivers know how to decode and dispatch it.
type MessageTypeID int

// Known message type identifiers. Each value corresponds to one of the
// message structs defined in this file.
const (
	// MessageTypeUnknown is the zero value and represents an unset or
	// unrecognized message type.
	MessageTypeUnknown MessageTypeID = iota
	// MessageTypeInitialize identifies an InitializeMessage.
	MessageTypeInitialize
	// MessageTypeEnclosureAirTemperature identifies an
	// EnclosureAirTemperatureMessage.
	MessageTypeEnclosureAirTemperature
	// MessageTypeEnclosureHumidity identifies an EnclosureHumidityMessage.
	MessageTypeEnclosureHumidity
	// MessageTypeSystemPH identifies a SystemPHMessage.
	MessageTypeSystemPH
	// MessageTypeSystemEC identifies a SystemECMessage.
	MessageTypeSystemEC
	// MessageTypeSystemOxygen identifies a SystemOxygenMessage.
	MessageTypeSystemOxygen
	// MessageTypeSystemWaterTemperature identifies a
	// SystemWaterTemperatureMessage.
	MessageTypeSystemWaterTemperature
	// MessageTypeSystemWaterProximity identifies a
	// SystemWaterProximityMessage.
	MessageTypeSystemWaterProximity
	// MessageTypeFlowRate identifies a FlowRateMessage.
	MessageTypeFlowRate
	// MessageTypePlantSiteLux identifies a PlantSiteLuxMessage.
	MessageTypePlantSiteLux
)

// InitializeMessage is sent when a device first connects to announce itself and
// establish a session with the coordinator. It currently carries no fields.
type InitializeMessage struct {
}

// EnclosureAirTemperatureMessage reports the air temperature measured inside a
// specific enclosure.
type EnclosureAirTemperatureMessage struct {
	// EnclosureID identifies the enclosure the reading came from.
	EnclosureID int `json:"enclosure_id"`
	// Temperature is the air temperature reading, in degrees Celsius.
	Temperature float64 `json:"temperature"`
}

// EnclosureHumidityMessage reports the relative humidity measured inside a
// specific enclosure.
type EnclosureHumidityMessage struct {
	// EnclosureID identifies the enclosure the reading came from.
	EnclosureID int `json:"enclosure_id"`
	// Humidity is the relative humidity reading, as a percentage.
	Humidity float64 `json:"humidity"`
}

// SystemPHMessage reports the pH level of the water in a system.
type SystemPHMessage struct {
	// SystemID identifies the system the reading came from.
	SystemID int `json:"system_id"`
	// PH is the measured pH of the system's water.
	PH float64 `json:"ph"`
}

// SystemECMessage reports the electrical conductivity of the water in a system,
// which is used as a proxy for nutrient concentration.
type SystemECMessage struct {
	// SystemID identifies the system the reading came from.
	SystemID int `json:"system_id"`
	// EC is the measured electrical conductivity.
	EC float64 `json:"ec"`
}

// SystemOxygenMessage reports the dissolved oxygen level of the water in a
// system.
type SystemOxygenMessage struct {
	// SystemID identifies the system the reading came from.
	SystemID int `json:"system_id"`
	// Oxygen is the measured dissolved oxygen level.
	Oxygen float64 `json:"oxygen"`
}

// SystemWaterTemperatureMessage reports the temperature of the water in a
// system.
type SystemWaterTemperatureMessage struct {
	// SystemID identifies the system the reading came from.
	SystemID int `json:"system_id"`
	// Temperature is the water temperature reading, in degrees Celsius.
	Temperature float64 `json:"temperature"`
}

// SystemWaterProximityMessage reports the water depth in a system as measured
// by a proximity sensor, which can be used to detect low-water conditions.
type SystemWaterProximityMessage struct {
	// SystemID identifies the system the reading came from.
	SystemID int `json:"system_id"`
	// Depth is the measured water depth.
	Depth float64 `json:"depth"`
}

// FlowRateMessage reports the volumetric flow rate through a specific flow
// sensor.
type FlowRateMessage struct {
	// FlowID identifies the flow sensor the reading came from.
	FlowID int `json:"flow_id"`
	// Rate is the measured flow rate.
	Rate float64 `json:"rate"`
}

// PlantSiteLuxMessage reports the light intensity measured at a specific plant
// site.
type PlantSiteLuxMessage struct {
	// PlantSiteID identifies the plant site the reading came from.
	PlantSiteID int `json:"plant_site_id"`
	// Lux is the measured illuminance, in lux.
	Lux int `json:"lux"`
}

// TimelapseImage holds a single still frame captured for a timelapse, along
// with the identifier of the camera that produced it.
type TimelapseImage struct {
	// MAC is the MAC address of the camera that captured the image.
	MAC string
	// Image is the raw encoded image data.
	Image []byte
}
