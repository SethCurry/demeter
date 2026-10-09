package httpd

import (
	"database/sql"
	"errors"
	"time"

	"github.com/SethCurry/demeter/internal/models"
	"github.com/SethCurry/demeter/internal/sensor"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	sqlite "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var upgrader = websocket.Upgrader{}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Time: time.Now(), Valid: true}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func toNullInt64(p *int64) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *p, Valid: true}
}

// toNullString converts an optional string pointer into a sql.NullString.
// A nil pointer (omitted or JSON null) becomes a NULL database value.
func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// toTime converts an optional time pointer into a non-null time.Time,
// defaulting to the current time when omitted. Used for note tables whose
// timestamp column is NOT NULL with no schema default.
func toTime(t *time.Time) time.Time {
	if t == nil {
		return time.Now()
	}
	return *t
}

// isConstraintViolation reports whether err is any SQLite constraint
// violation (foreign key, unique, not null, ...). The driver surfaces the
// extended result codes (e.g. SQLITE_CONSTRAINT_UNIQUE), so the comparison
// masks off the extended bits and checks the primary SQLITE_CONSTRAINT code.
func isConstraintViolation(err error) bool {
	var sqliteErr *sqlite.Error
	return errors.As(err, &sqliteErr) && sqliteErr.Code()&0xff == sqlite3.SQLITE_CONSTRAINT
}

// isUniqueViolation reports whether err is a SQLite UNIQUE constraint
// violation, such as inserting a duplicate value in a unique column.
func isUniqueViolation(err error) bool {
	var sqliteErr *sqlite.Error
	return errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE
}

func createSensorConnectionHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		w, r := ctx.Writer, ctx.Request

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		receiver := sensor.NewSensorReceiver(db)
		conn := sensor.NewConn(c, receiver)

		conn.Run()

		ctx.JSON(200, map[string]string{"message": "connected"})
	}
}

func Run(addr string, db *models.Queries) error {
	r := gin.Default()
	r.GET("/api/websocket/sensors", createSensorConnectionHandler(db))

	enclosures := r.Group("/api/enclosures")
	{
		enclosures.GET("", listEnclosuresHandler(db))
		enclosures.POST("", createEnclosureHandler(db))
		enclosures.GET("/:id", getEnclosureHandler(db))
		enclosures.GET("/:id/air-temperature", listEnclosureAirTemperaturesHandler(db))
		enclosures.GET("/:id/air-humidity", listEnclosureAirHumidityHandler(db))
		enclosures.GET("/:id/plants", listEnclosurePlantsHandler(db))
		enclosures.PUT("/:id", updateEnclosureHandler(db))
		enclosures.DELETE("/:id", deleteEnclosureHandler(db))
	}

	systems := r.Group("/api/systems")
	{
		systems.GET("", listSystemsHandler(db))
		systems.POST("", createSystemHandler(db))
		systems.GET("/:id", getSystemHandler(db))
		systems.GET("/:id/ph", listSystemPHHandler(db))
		systems.GET("/:id/ec", listSystemECHandler(db))
		systems.GET("/:id/water-temperature", listSystemWaterTemperatureHandler(db))
		systems.PUT("/:id", updateSystemHandler(db))
		systems.DELETE("/:id", deleteSystemHandler(db))
	}

	flows := r.Group("/api/flows")
	{
		flows.GET("", listFlowsHandler(db))
		flows.POST("", createFlowHandler(db))
		flows.GET("/:id", getFlowHandler(db))
		flows.PUT("/:id", updateFlowHandler(db))
		flows.DELETE("/:id", deleteFlowHandler(db))
		flows.GET("/:id/plants", listFlowPlantsHandler(db))
	}

	plantSites := r.Group("/api/plant-sites")
	{
		plantSites.GET("", listPlantSitesHandler(db))
		plantSites.POST("", createPlantSiteHandler(db))
		plantSites.GET("/:id", getPlantSiteHandler(db))
		plantSites.PUT("/:id", updatePlantSiteHandler(db))
		plantSites.DELETE("/:id", deletePlantSiteHandler(db))
	}

	plants := r.Group("/api/plants")
	{
		plants.GET("", listPlantsHandler(db))
		plants.POST("", createPlantHandler(db))
		plants.GET("/:id", getPlantHandler(db))
		plants.PUT("/:id", updatePlantHandler(db))
		plants.DELETE("/:id", deletePlantHandler(db))
	}

	plantGenera := r.Group("/api/plant-genera")
	{
		plantGenera.GET("", listPlantGeneraHandler(db))
		plantGenera.POST("", createPlantGenusHandler(db))
		plantGenera.GET("/:id", getPlantGenusHandler(db))
		plantGenera.PUT("/:id", updatePlantGenusHandler(db))
		plantGenera.DELETE("/:id", deletePlantGenusHandler(db))
	}

	plantSpecies := r.Group("/api/plant-species")
	{
		plantSpecies.GET("", listPlantSpeciesHandler(db))
		plantSpecies.POST("", createPlantSpeciesHandler(db))
		plantSpecies.GET("/:id", getPlantSpeciesHandler(db))
		plantSpecies.PUT("/:id", updatePlantSpeciesHandler(db))
		plantSpecies.DELETE("/:id", deletePlantSpeciesHandler(db))
	}

	enclosureNotes := r.Group("/api/enclosure-notes")
	{
		enclosureNotes.GET("", listSimpleNotesHandler("enclosure note", db.ListEnclosureNotes))
		enclosureNotes.POST("", createSimpleNoteHandler("enclosure note",
			func(req simpleNoteRequest) models.CreateEnclosureNoteParams {
				return models.CreateEnclosureNoteParams{
					Timestamp: toNullTime(req.Timestamp),
					Content:   toNullString(req.Content),
				}
			}, db.CreateEnclosureNote))
		enclosureNotes.GET("/:id", getSimpleNoteHandler("enclosure note", db.GetEnclosureNote))
		enclosureNotes.PUT("/:id", updateSimpleNoteHandler("enclosure note",
			func(req simpleNoteRequest, id int64) models.UpdateEnclosureNoteParams {
				return models.UpdateEnclosureNoteParams{
					Timestamp: toNullTime(req.Timestamp),
					Content:   toNullString(req.Content),
					ID:        id,
				}
			}, db.UpdateEnclosureNote))
		enclosureNotes.DELETE("/:id", deleteSimpleNoteHandler("enclosure note", db.DeleteEnclosureNote))
	}

	systemNotes := r.Group("/api/system-notes")
	{
		systemNotes.GET("", listSimpleNotesHandler("system note", db.ListSystemNotes))
		systemNotes.POST("", createSimpleNoteHandler("system note",
			func(req simpleNoteRequest) models.CreateSystemNoteParams {
				return models.CreateSystemNoteParams{
					Timestamp: toNullTime(req.Timestamp),
					Content:   toNullString(req.Content),
				}
			}, db.CreateSystemNote))
		systemNotes.GET("/:id", getSimpleNoteHandler("system note", db.GetSystemNote))
		systemNotes.PUT("/:id", updateSimpleNoteHandler("system note",
			func(req simpleNoteRequest, id int64) models.UpdateSystemNoteParams {
				return models.UpdateSystemNoteParams{
					Timestamp: toNullTime(req.Timestamp),
					Content:   toNullString(req.Content),
					ID:        id,
				}
			}, db.UpdateSystemNote))
		systemNotes.DELETE("/:id", deleteSimpleNoteHandler("system note", db.DeleteSystemNote))
	}

	flowNotes := r.Group("/api/flow-notes")
	{
		flowNotes.GET("", listSimpleNotesHandler("flow note", db.ListFlowNotes))
		flowNotes.POST("", createSimpleNoteHandler("flow note",
			func(req simpleNoteRequest) models.CreateFlowNoteParams {
				return models.CreateFlowNoteParams{
					Timestamp: toNullTime(req.Timestamp),
					Content:   toNullString(req.Content),
				}
			}, db.CreateFlowNote))
		flowNotes.GET("/:id", getSimpleNoteHandler("flow note", db.GetFlowNote))
		flowNotes.PUT("/:id", updateSimpleNoteHandler("flow note",
			func(req simpleNoteRequest, id int64) models.UpdateFlowNoteParams {
				return models.UpdateFlowNoteParams{
					Timestamp: toNullTime(req.Timestamp),
					Content:   toNullString(req.Content),
					ID:        id,
				}
			}, db.UpdateFlowNote))
		flowNotes.DELETE("/:id", deleteSimpleNoteHandler("flow note", db.DeleteFlowNote))
	}

	plantSiteNotes := r.Group("/api/plant-site-notes")
	{
		plantSiteNotes.GET("", listPlantSiteNotesHandler(db))
		plantSiteNotes.POST("", createPlantSiteNoteHandler(db))
		plantSiteNotes.GET("/:id", getPlantSiteNoteHandler(db))
		plantSiteNotes.PUT("/:id", updatePlantSiteNoteHandler(db))
		plantSiteNotes.DELETE("/:id", deletePlantSiteNoteHandler(db))
	}

	plantNotes := r.Group("/api/plant-notes")
	{
		plantNotes.GET("", listPlantNotesHandler(db))
		plantNotes.POST("", createPlantNoteHandler(db))
		plantNotes.GET("/:id", getPlantNoteHandler(db))
		plantNotes.PUT("/:id", updatePlantNoteHandler(db))
		plantNotes.DELETE("/:id", deletePlantNoteHandler(db))
	}

	return r.Run(addr)
}
