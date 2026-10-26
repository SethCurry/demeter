package sensor

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/SethCurry/demeter/internal/models"
	"github.com/SethCurry/demeter/pkg/demeter"
	"github.com/rs/zerolog/log"
)

// NewSensorReceiver creates a new SensorReceiver that persists sensor readings
// using the given database queries and stores timelapse images under the
// default timelapse directory.
func NewSensorReceiver(db *models.Queries) *SensorReceiver {
	return &SensorReceiver{db: db, timelapseDir: "/srv/demeter/timelapse"}
}

type SensorReceiver struct {
	db *models.Queries
	timelapseDir string
}

// Initialize handles an initialize message sent by a sensor when it connects.
// It currently only logs the event.
func (r *SensorReceiver) Initialize(initMsg demeter.InitializeMessage) {
	log.Info().Msg("received initialize message")
}

// EnclosureAirTemperature records the air temperature reported for an
// enclosure. It logs an error if the measurement cannot be stored.
func (r *SensorReceiver) EnclosureAirTemperature(airTempMsg demeter.EnclosureAirTemperatureMessage) {
	log.Info().Int("enclosure_id", airTempMsg.EnclosureID).Float64("temperature_c", airTempMsg.Temperature).Msg("received enclosure air temperature message")
	err := r.db.CreateEnclosureAirTemperature(context.Background(), models.CreateEnclosureAirTemperatureParams{
		EnclosureID:  int64(airTempMsg.EnclosureID),
		TemperatureC: airTempMsg.Temperature,
	})
	if err != nil {
		log.Error().Err(err).Int("enclosure_id", airTempMsg.EnclosureID).Msg("failed to record enclosure air temperature")
	}
}

// EnclosureHumidity records the relative humidity reported for an enclosure.
// It logs an error if the measurement cannot be stored.
func (r *SensorReceiver) EnclosureHumidity(humidityMsg demeter.EnclosureHumidityMessage) {
	log.Info().Int("enclosure_id", humidityMsg.EnclosureID).Float64("humidity_rh", humidityMsg.Humidity).Msg("received enclosure humidity message")
	err := r.db.CreateEnclosureAirHumidity(context.Background(), models.CreateEnclosureAirHumidityParams{
		EnclosureID: int64(humidityMsg.EnclosureID),
		HumidityRh:  humidityMsg.Humidity,
	})
	if err != nil {
		log.Error().Err(err).Int("enclosure_id", humidityMsg.EnclosureID).Msg("failed to record enclosure air humidity")
	}
}

// SystemPH records the pH level reported for a system. It logs an error if the
// measurement cannot be stored.
func (r *SensorReceiver) SystemPH(phMsg demeter.SystemPHMessage) {
	log.Info().Int("system_id", phMsg.SystemID).Float64("ph", phMsg.PH).Msg("received system pH message")
	err := r.db.CreateSystemPH(context.Background(), models.CreateSystemPHParams{
		SystemID: int64(phMsg.SystemID),
		Ph:       phMsg.PH,
	})
	if err != nil {
		log.Error().Err(err).Int("system_id", phMsg.SystemID).Msg("failed to record system pH")
	}
}

// SystemEC records the electrical conductivity reported for a system. It logs
// an error if the measurement cannot be stored.
func (r *SensorReceiver) SystemEC(ecMsg demeter.SystemECMessage) {
	log.Info().Int("system_id", ecMsg.SystemID).Float64("ec", ecMsg.EC).Msg("received system EC message")
	err := r.db.CreateSystemEC(context.Background(), models.CreateSystemECParams{
		SystemID: int64(ecMsg.SystemID),
		Ec:       ecMsg.EC,
	})
	if err != nil {
		log.Error().Err(err).Int("system_id", ecMsg.SystemID).Msg("failed to record system EC")
	}
}

// SystemOxygen records the dissolved oxygen level reported for a system. It
// logs an error if the measurement cannot be stored.
func (r *SensorReceiver) SystemOxygen(oxygenMsg demeter.SystemOxygenMessage) {
	log.Info().Int("system_id", oxygenMsg.SystemID).Float64("oxygen", oxygenMsg.Oxygen).Msg("received system oxygen message")
	err := r.db.CreateSystemOxygen(context.Background(), models.CreateSystemOxygenParams{
		SystemID: int64(oxygenMsg.SystemID),
		Oxygen:   oxygenMsg.Oxygen,
	})
	if err != nil {
		log.Error().Err(err).Int("system_id", oxygenMsg.SystemID).Msg("failed to record system oxygen")
	}
}

// SystemWaterTemperature records the water temperature reported for a system.
// It logs an error if the measurement cannot be stored.
func (r *SensorReceiver) SystemWaterTemperature(waterTempMsg demeter.SystemWaterTemperatureMessage) {
	log.Info().Int("system_id", waterTempMsg.SystemID).Float64("temperature_c", waterTempMsg.Temperature).Msg("received system water temperature message")
	err := r.db.CreateSystemWaterTemperature(context.Background(), models.CreateSystemWaterTemperatureParams{
		SystemID:          int64(waterTempMsg.SystemID),
		WaterTemperatureC: waterTempMsg.Temperature,
	})
	if err != nil {
		log.Error().Err(err).Int("system_id", waterTempMsg.SystemID).Msg("failed to record system water temperature")
	}
}

// SystemWaterProximity records the water proximity (depth) reported for a
// system. It logs an error if the measurement cannot be stored.
func (r *SensorReceiver) SystemWaterProximity(waterProximityMsg demeter.SystemWaterProximityMessage) {
	log.Info().Int("system_id", waterProximityMsg.SystemID).Float64("depth_m", waterProximityMsg.Depth).Msg("received system water proximity message")
	err := r.db.CreateSystemWaterProximity(context.Background(), models.CreateSystemWaterProximityParams{
		SystemID:   int64(waterProximityMsg.SystemID),
		ProximityM: waterProximityMsg.Depth,
	})
	if err != nil {
		log.Error().Err(err).Int("system_id", waterProximityMsg.SystemID).Msg("failed to record system water proximity")
	}
}

// FlowRate records the flow rate reported by a flow sensor. It logs an error
// if the measurement cannot be stored.
func (r *SensorReceiver) FlowRate(flowRateMsg demeter.FlowRateMessage) {
	log.Info().Int("flow_id", flowRateMsg.FlowID).Float64("rate_mps", flowRateMsg.Rate).Msg("received flow rate message")
	err := r.db.CreateFlowRate(context.Background(), models.CreateFlowRateParams{
		FlowID:  int64(flowRateMsg.FlowID),
		RateMps: flowRateMsg.Rate,
	})
	if err != nil {
		log.Error().Err(err).Int("flow_id", flowRateMsg.FlowID).Msg("failed to record flow rate")
	}
}

// PlantSiteLux records the light level reported for a plant site. It logs an
// error if the measurement cannot be stored.
func (r *SensorReceiver) PlantSiteLux(plantSiteLuxMsg demeter.PlantSiteLuxMessage) {
	log.Info().Int("plant_site_id", plantSiteLuxMsg.PlantSiteID).Int("lux", plantSiteLuxMsg.Lux).Msg("received plant site lux message")
	err := r.db.CreatePlantSiteLux(context.Background(), models.CreatePlantSiteLuxParams{
		PlantSiteID: int64(plantSiteLuxMsg.PlantSiteID),
		Lux:         int64(plantSiteLuxMsg.Lux),
	})
	if err != nil {
		log.Error().Err(err).Int("plant_site_id", plantSiteLuxMsg.PlantSiteID).Msg("failed to record plant site lux")
	}
}

// Timelapse saves a timelapse image to disk, storing it in a per-device
// directory keyed by MAC address. Errors creating the directory or file are
// not returned; a failed image write is silently skipped.
func (r *SensorReceiver) Timelapse(msg demeter.TimelapseImage) {
	log.Info().Str("mac", msg.MAC).Int("image_bytes", len(msg.Image)).Msg("received timelapse image message")
	deviceDir := filepath.Join(r.timelapseDir, msg.MAC)

	if _, err := os.Stat(deviceDir); err != nil && os.IsNotExist(err) {
		log.Info().Str("path", deviceDir).Msg("creating directory for timelapses")
		os.MkdirAll(deviceDir, 0755)
	}

	currentTime := time.Now()
	imagePath := filepath.Join(deviceDir, currentTime.Format(time.RFC3339) + ".jpeg")

	fd, err := os.Create(imagePath)
	if err != nil {
		return
	}
	defer fd.Close()

	fd.Write(msg.Image)
	log.Info().Str("path", imagePath).Msg("saved timelapse image")
}
