package sensor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/SethCurry/demeter/internal/models"
	"github.com/SethCurry/demeter/pkg/demeter"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

func NewSensorReceiver(db *models.Queries) *SensorReceiver {
	return &SensorReceiver{db: db}
}

type SensorReceiver struct {
	db *models.Queries
}

func (r *SensorReceiver) Initialize(initMsg demeter.InitializeMessage) {

}

func (r *SensorReceiver) EnclosureAirTemperature(airTempMsg demeter.EnclosureAirTemperatureMessage) {
	err := r.db.CreateEnclosureAirTemperature(context.Background(), models.CreateEnclosureAirTemperatureParams{
		EnclosureID:  int64(airTempMsg.EnclosureID),
		TemperatureC: airTempMsg.Temperature,
	})
	if err != nil {
		log.Error().Err(err).Int("enclosure_id", airTempMsg.EnclosureID).Msg("failed to record enclosure air temperature")
	}
}

func (r *SensorReceiver) EnclosureHumidity(humidityMsg demeter.EnclosureHumidityMessage) {
	err := r.db.CreateEnclosureAirHumidity(context.Background(), models.CreateEnclosureAirHumidityParams{
		EnclosureID: int64(humidityMsg.EnclosureID),
		HumidityRh:  humidityMsg.Humidity,
	})
	if err != nil {
		log.Error().Err(err).Int("enclosure_id", humidityMsg.EnclosureID).Msg("failed to record enclosure air humidity")
	}
}

func (r *SensorReceiver) SystemPH(phMsg demeter.SystemPHMessage) {
	err := r.db.CreateSystemPH(context.Background(), models.CreateSystemPHParams{
		SystemID: int64(phMsg.SystemID),
		Ph:       phMsg.PH,
	})
	if err != nil {
		log.Error().Err(err).Int("system_id", phMsg.SystemID).Msg("failed to record system pH")
	}
}

func (r *SensorReceiver) SystemEC(ecMsg demeter.SystemECMessage) {
	err := r.db.CreateSystemEC(context.Background(), models.CreateSystemECParams{
		SystemID: int64(ecMsg.SystemID),
		Ec:       ecMsg.EC,
	})
	if err != nil {
		log.Error().Err(err).Int("system_id", ecMsg.SystemID).Msg("failed to record system EC")
	}
}

func (r *SensorReceiver) SystemOxygen(oxygenMsg demeter.SystemOxygenMessage) {
	err := r.db.CreateSystemOxygen(context.Background(), models.CreateSystemOxygenParams{
		SystemID: int64(oxygenMsg.SystemID),
		Oxygen:   oxygenMsg.Oxygen,
	})
	if err != nil {
		log.Error().Err(err).Int("system_id", oxygenMsg.SystemID).Msg("failed to record system oxygen")
	}
}

func (r *SensorReceiver) SystemWaterTemperature(waterTempMsg demeter.SystemWaterTemperatureMessage) {
	err := r.db.CreateSystemWaterTemperature(context.Background(), models.CreateSystemWaterTemperatureParams{
		SystemID:          int64(waterTempMsg.SystemID),
		WaterTemperatureC: waterTempMsg.Temperature,
	})
	if err != nil {
		log.Error().Err(err).Int("system_id", waterTempMsg.SystemID).Msg("failed to record system water temperature")
	}
}

func (r *SensorReceiver) SystemWaterProximity(waterProximityMsg demeter.SystemWaterProximityMessage) {
	err := r.db.CreateSystemWaterProximity(context.Background(), models.CreateSystemWaterProximityParams{
		SystemID:   int64(waterProximityMsg.SystemID),
		ProximityM: waterProximityMsg.Depth,
	})
	if err != nil {
		log.Error().Err(err).Int("system_id", waterProximityMsg.SystemID).Msg("failed to record system water proximity")
	}
}

func (r *SensorReceiver) FlowRate(flowRateMsg demeter.FlowRateMessage) {
	err := r.db.CreateFlowRate(context.Background(), models.CreateFlowRateParams{
		FlowID:  int64(flowRateMsg.FlowID),
		RateMps: flowRateMsg.Rate,
	})
	if err != nil {
		log.Error().Err(err).Int("flow_id", flowRateMsg.FlowID).Msg("failed to record flow rate")
	}
}

func (r *SensorReceiver) PlantSiteLux(plantSiteLuxMsg demeter.PlantSiteLuxMessage) {
	err := r.db.CreatePlantSiteLux(context.Background(), models.CreatePlantSiteLuxParams{
		PlantSiteID: int64(plantSiteLuxMsg.PlantSiteID),
		Lux:         int64(plantSiteLuxMsg.Lux),
	})
	if err != nil {
		log.Error().Err(err).Int("plant_site_id", plantSiteLuxMsg.PlantSiteID).Msg("failed to record plant site lux")
	}
}

func NewConn(ws *websocket.Conn, receiver *SensorReceiver) *Conn {
	return &Conn{ws: ws, receiver: receiver}
}

type Conn struct {
	ws       *websocket.Conn
	receiver *SensorReceiver
}

func (c *Conn) Run() error {
	for {
		_, messageData, err := c.ws.ReadMessage()
		if err != nil {
			return err
		}

		asMsg := string(messageData)

		msgParts := strings.Split(asMsg, ":")
		if len(msgParts) < 3 {
			return fmt.Errorf("invalid message format, expected at least 3 parts: %q", asMsg)
		}

		messageTypeIDInt, err := strconv.Atoi(msgParts[0])
		if err != nil {
			return fmt.Errorf("invalid message type ID %q: %w", msgParts[0], err)
		}
		messageTypeID := demeter.MessageTypeID(messageTypeIDInt)

		targetID, err := strconv.Atoi(msgParts[1])
		if err != nil {
			return fmt.Errorf("invalid target ID %q: %w", msgParts[1], err)
		}

		valueStr := msgParts[2]

		switch messageTypeID {
		case demeter.MessageTypeInitialize:
			c.receiver.Initialize(demeter.InitializeMessage{})
		case demeter.MessageTypeEnclosureAirTemperature:
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid measurement value %q: %w", valueStr, err)
			}
			c.receiver.EnclosureAirTemperature(demeter.EnclosureAirTemperatureMessage{
				EnclosureID: targetID,
				Temperature: float64(value) / 1000,
			})
		case demeter.MessageTypeEnclosureHumidity:
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid measurement value %q: %w", valueStr, err)
			}
			c.receiver.EnclosureHumidity(demeter.EnclosureHumidityMessage{
				EnclosureID: targetID,
				Humidity:    float64(value) / 1000,
			})
		case demeter.MessageTypeSystemPH:
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid measurement value %q: %w", valueStr, err)
			}
			c.receiver.SystemPH(demeter.SystemPHMessage{
				SystemID: targetID,
				PH:       float64(value) / 1000,
			})
		case demeter.MessageTypeSystemEC:
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid measurement value %q: %w", valueStr, err)
			}
			c.receiver.SystemEC(demeter.SystemECMessage{
				SystemID: targetID,
				EC:       float64(value) / 1000,
			})
		case demeter.MessageTypeSystemOxygen:
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid measurement value %q: %w", valueStr, err)
			}
			c.receiver.SystemOxygen(demeter.SystemOxygenMessage{
				SystemID: targetID,
				Oxygen:   float64(value) / 1000,
			})
		case demeter.MessageTypeSystemWaterTemperature:
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid measurement value %q: %w", valueStr, err)
			}
			c.receiver.SystemWaterTemperature(demeter.SystemWaterTemperatureMessage{
				SystemID:    targetID,
				Temperature: float64(value) / 1000,
			})
		case demeter.MessageTypeSystemWaterProximity:
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid measurement value %q: %w", valueStr, err)
			}
			c.receiver.SystemWaterProximity(demeter.SystemWaterProximityMessage{
				SystemID: targetID,
				Depth:    float64(value) / 1000,
			})
		case demeter.MessageTypeFlowRate:
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid measurement value %q: %w", valueStr, err)
			}
			c.receiver.FlowRate(demeter.FlowRateMessage{
				FlowID: targetID,
				Rate:   float64(value) / 1000,
			})
		case demeter.MessageTypePlantSiteLux:
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid measurement value %q: %w", valueStr, err)
			}
			c.receiver.PlantSiteLux(demeter.PlantSiteLuxMessage{
				PlantSiteID: targetID,
				Lux:         int(value),
			})
		default:
			log.Warn().Int("message_type_id", int(messageTypeID)).Msg("unknown message type")
		}
	}
}
