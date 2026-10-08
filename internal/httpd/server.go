package httpd

import (
	"database/sql"
	"time"

	"github.com/SethCurry/demeter/internal/models"
	"github.com/SethCurry/demeter/internal/sensor"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type enclosureRequest struct {
	Name string `json:"name" binding:"required"`
}

type systemRequest struct {
	Name        string `json:"name" binding:"required"`
	EnclosureID int64  `json:"enclosure_id" binding:"required"`
}

type flowRequest struct {
	Name         string `json:"name" binding:"required"`
	SystemID     int64  `json:"system_id" binding:"required"`
	ParentFlowID *int64 `json:"parent_flow_id,omitempty"`
}

type plantSiteRequest struct {
	FlowID int64 `json:"flow_id" binding:"required"`
	X      int64 `json:"x" binding:"required"`
	Y      int64 `json:"y" binding:"required"`
	Z      int64 `json:"z" binding:"required"`
}

type plantRequest struct {
	PlantSiteID int64      `json:"plant_site_id" binding:"required"`
	PlantedOn   *time.Time `json:"planted_on,omitempty"`
}

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
		enclosures.PUT("/:id", updateEnclosureHandler(db))
		enclosures.DELETE("/:id", deleteEnclosureHandler(db))
	}

	systems := r.Group("/api/systems")
	{
		systems.GET("", listSystemsHandler(db))
		systems.POST("", createSystemHandler(db))
		systems.GET("/:id", getSystemHandler(db))
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
