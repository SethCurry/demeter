package demeter

type MessageTypeID int

const (
	MessageTypeUnknown MessageTypeID = iota
	MessageTypeInitialize
	MessageTypeEnclosureAirTemperature
	MessageTypeEnclosureHumidity
	MessageTypeSystemPH
	MessageTypeSystemEC
	MessageTypeSystemOxygen
	MessageTypeSystemWaterTemperature
	MessageTypeSystemWaterProximity
	MessageTypeFlowRate
	MessageTypePlantSiteLux
)

type InitializeMessage struct {
}

type EnclosureAirTemperatureMessage struct {
	EnclosureID int     `json:"enclosure_id"`
	Temperature float64 `json:"temperature"`
}

type EnclosureHumidityMessage struct {
	EnclosureID int     `json:"enclosure_id"`
	Humidity    float64 `json:"humidity"`
}

type SystemPHMessage struct {
	SystemID int     `json:"system_id"`
	PH       float64 `json:"ph"`
}

type SystemECMessage struct {
	SystemID int     `json:"system_id"`
	EC       float64 `json:"ec"`
}

type SystemOxygenMessage struct {
	SystemID int     `json:"system_id"`
	Oxygen   float64 `json:"oxygen"`
}

type SystemWaterTemperatureMessage struct {
	SystemID    int     `json:"system_id"`
	Temperature float64 `json:"temperature"`
}

type SystemWaterProximityMessage struct {
	SystemID int     `json:"system_id"`
	Depth    float64 `json:"depth"`
}

type FlowRateMessage struct {
	FlowID int     `json:"flow_id"`
	Rate   float64 `json:"rate"`
}

type PlantSiteLuxMessage struct {
	PlantSiteID int `json:"plant_site_id"`
	Lux         int `json:"lux"`
}
