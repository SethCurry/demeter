package sensor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/SethCurry/demeter/pkg/demeter"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// NewConn creates a new Conn that reads sensor data from the given websocket
// and forwards it to the provided SensorReceiver.
func NewConn(ws *websocket.Conn, receiver *SensorReceiver) *Conn {
	return &Conn{ws: ws, receiver: receiver}
}

type Conn struct {
	ws       *websocket.Conn
	receiver *SensorReceiver
}

// Run reads messages from the websocket until the connection closes or an
// error occurs. Text messages are parsed as colon-delimited sensor messages
// and dispatched to the receiver, while binary messages are treated as
// timelapse images. It returns nil after successfully handling an image and
// returns an error if the connection fails or a message cannot be parsed.
func (c *Conn) Run() error {
	for {
		wsMessageType, messageData, err := c.ws.ReadMessage()
		if err != nil {
			return err
		}

		log.Info().Int("message_type", wsMessageType).Msg("received websocket message")

		if wsMessageType == websocket.BinaryMessage {
			macAddr := messageData[:12]
			imageData := messageData[12:]
			c.receiver.Timelapse(demeter.TimelapseImage{
				MAC: string(macAddr),
				Image: imageData,
			})
			return nil
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
