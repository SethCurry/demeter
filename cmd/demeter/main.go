package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"math/rand/v2"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/SethCurry/demeter/internal/httpd"
	"github.com/SethCurry/demeter/internal/models"
	"github.com/SethCurry/demeter/pkg/demeter"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

type SQLConfig struct {
	Path string
}

type HTTPConfig struct {
	Addr string
}

type Config struct {
	HTTP HTTPConfig
	SQL  SQLConfig
}

// ensureSchema applies the embedded schema.sql to the database so that all
// tables exist before the server starts.
// floatMsg formats a float measurement for the wire protocol. The client is
// expected to multiply float values by 1000 and send them as integers.
func floatMsg(typeID demeter.MessageTypeID, targetID int, value float64) string {
	return strconv.Itoa(int(typeID)) + ":" + strconv.Itoa(targetID) + ":" + strconv.Itoa(int(value*1000))
}

// intMsg formats an integer measurement for the wire protocol.
func intMsg(typeID demeter.MessageTypeID, targetID int, value int) string {
	return strconv.Itoa(int(typeID)) + ":" + strconv.Itoa(targetID) + ":" + strconv.Itoa(value)
}

// jitter returns center plus a random offset in [-amp, amp).
func jitter(center, amp float64) float64 {
	return center + (rand.Float64()*2-1)*amp
}

func runSensorSimulator(ctx context.Context, cmd *cli.Command) error {
	addr := cmd.String("addr")
	interval := cmd.Duration("interval")

	u := url.URL{Scheme: "ws", Host: addr, Path: "/api/websocket/sensors"}
	log.Info().Str("url", u.String()).Msg("connecting to websocket")

	c, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial websocket: %w", err)
	}
	defer c.Close()

	// Send an initialize message first.
	if err := c.WriteMessage(websocket.TextMessage, []byte(intMsg(demeter.MessageTypeInitialize, 0, 0))); err != nil {
		return fmt.Errorf("send initialize: %w", err)
	}

	const targetID = 1

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		messages := []string{
			floatMsg(demeter.MessageTypeEnclosureAirTemperature, targetID, jitter(22.5, 1)),
			floatMsg(demeter.MessageTypeEnclosureHumidity, targetID, jitter(65, 5)),
			floatMsg(demeter.MessageTypeSystemPH, targetID, jitter(6.5, 0.2)),
			floatMsg(demeter.MessageTypeSystemEC, targetID, jitter(1.2, 0.1)),
			floatMsg(demeter.MessageTypeSystemOxygen, targetID, jitter(8, 0.5)),
			floatMsg(demeter.MessageTypeSystemWaterTemperature, targetID, jitter(18.5, 1)),
			floatMsg(demeter.MessageTypeSystemWaterProximity, targetID, jitter(0.5, 0.05)),
			floatMsg(demeter.MessageTypeFlowRate, targetID, jitter(2, 0.2)),
			intMsg(demeter.MessageTypePlantSiteLux, targetID, 12000+rand.IntN(6000)),
		}

		for _, msg := range messages {
			log.Debug().Str("message", msg).Msg("sending")
			if err := c.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				return fmt.Errorf("write message: %w", err)
			}
		}

		log.Info().Int("count", len(messages)).Msg("sent synthetic readings")
	}
}

func ensureSchema(db *sql.DB) error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("read embedded schema: %w", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		return nil
		//return fmt.Errorf("apply schema: %w", err)
	}

	return nil
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg := &Config{
		HTTP: HTTPConfig{
			Addr: ":8249",
		},
		SQL: SQLConfig{
			Path: "./demeter.sqlite",
		},
	}

	db, err := sql.Open("sqlite", cfg.SQL.Path)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}
	defer db.Close()

	if err := ensureSchema(db); err != nil {
		log.Fatal().Err(err).Msg("failed to apply database schema")
	}

	queries := models.New(db)

	cmd := &cli.Command{
		Name:  "demeter",
		Usage: "Demeter server",
		Commands: []*cli.Command{
			{
				Name:  "server",
				Usage: "Run the Demeter server",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return httpd.Run(cfg.HTTP.Addr, queries)
				},
			},
			{
				Name:  "simulate-sensor",
				Usage: "Pretend to be a websocket sensor and send synthetic data",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "addr",
						Usage:   "Address of the Demeter server",
						Value:   "localhost:8080",
						Sources: cli.EnvVars("DEMETER_ADDR"),
					},
					&cli.DurationFlag{
						Name:  "interval",
						Usage: "Interval between sending each batch of readings",
						Value: 5 * time.Second,
					},
				},
				Action: runSensorSimulator,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal().Err(err).Msg("error running command")
	}
}
