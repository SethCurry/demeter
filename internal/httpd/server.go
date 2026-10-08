package httpd

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/SethCurry/demeter/internal/models"
	"github.com/SethCurry/demeter/internal/sensor"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	sqlite "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
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

func listEnclosuresHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		enclosures, err := db.ListEnclosures(ctx.Request.Context())
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, enclosures)
	}
}

func getEnclosureHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		enclosure, err := db.GetEnclosure(ctx.Request.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "enclosure not found"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, enclosure)
	}
}

func createEnclosureHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req enclosureRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		enclosure, err := db.CreateEnclosure(ctx.Request.Context(), req.Name)
		if err != nil {
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "enclosure name already exists"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusCreated, enclosure)
	}
}

func updateEnclosureHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req enclosureRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		enclosure, err := db.UpdateEnclosure(ctx.Request.Context(), models.UpdateEnclosureParams{
			Name: req.Name,
			ID:   id,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "enclosure not found"})
				return
			}
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "enclosure name already exists"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, enclosure)
	}
}

func deleteEnclosureHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := db.DeleteEnclosure(ctx.Request.Context(), id); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.Status(http.StatusNoContent)
	}
}

func listSystemsHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var systems []models.System
		var err error
		if enclosureIDStr := ctx.Query("enclosure_id"); enclosureIDStr != "" {
			enclosureID, parseErr := strconv.ParseInt(enclosureIDStr, 10, 64)
			if parseErr != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid enclosure_id"})
				return
			}
			systems, err = db.ListSystemsByEnclosure(ctx.Request.Context(), enclosureID)
		} else {
			systems, err = db.ListSystems(ctx.Request.Context())
		}
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, systems)
	}
}

func getSystemHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		system, err := db.GetSystem(ctx.Request.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "system not found"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, system)
	}
}

func createSystemHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req systemRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		system, err := db.CreateSystem(ctx.Request.Context(), models.CreateSystemParams{
			Name:        req.Name,
			EnclosureID: req.EnclosureID,
		})
		if err != nil {
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "system name already exists or enclosure does not exist"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusCreated, system)
	}
}

func updateSystemHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req systemRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		system, err := db.UpdateSystem(ctx.Request.Context(), models.UpdateSystemParams{
			Name:        req.Name,
			EnclosureID: req.EnclosureID,
			ID:          id,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "system not found"})
				return
			}
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "system name already exists or enclosure does not exist"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, system)
	}
}

func deleteSystemHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := db.DeleteSystem(ctx.Request.Context(), id); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.Status(http.StatusNoContent)
	}
}

func listFlowsHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var flows []models.Flow
		var err error
		systemIDStr := ctx.Query("system_id")
		parentFlowIDStr := ctx.Query("parent_flow_id")
		switch {
		case systemIDStr != "":
			systemID, parseErr := strconv.ParseInt(systemIDStr, 10, 64)
			if parseErr != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid system_id"})
				return
			}
			flows, err = db.ListFlowsBySystem(ctx.Request.Context(), systemID)
		case parentFlowIDStr != "":
			parentFlowID, parseErr := strconv.ParseInt(parentFlowIDStr, 10, 64)
			if parseErr != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid parent_flow_id"})
				return
			}
			flows, err = db.ListChildFlows(ctx.Request.Context(), sql.NullInt64{Int64: parentFlowID, Valid: true})
		default:
			flows, err = db.ListFlows(ctx.Request.Context())
		}
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, flows)
	}
}

func getFlowHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		flow, err := db.GetFlow(ctx.Request.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "flow not found"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, flow)
	}
}

func createFlowHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req flowRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		flow, err := db.CreateFlow(ctx.Request.Context(), models.CreateFlowParams{
			Name:         req.Name,
			SystemID:     req.SystemID,
			ParentFlowID: toNullInt64(req.ParentFlowID),
		})
		if err != nil {
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "flow name already exists or referenced system/parent flow does not exist"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusCreated, flow)
	}
}

func updateFlowHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req flowRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		flow, err := db.UpdateFlow(ctx.Request.Context(), models.UpdateFlowParams{
			Name:         req.Name,
			SystemID:     req.SystemID,
			ParentFlowID: toNullInt64(req.ParentFlowID),
			ID:           id,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "flow not found"})
				return
			}
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "flow name already exists or referenced system/parent flow does not exist"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, flow)
	}
}

func deleteFlowHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := db.DeleteFlow(ctx.Request.Context(), id); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.Status(http.StatusNoContent)
	}
}

func listPlantSitesHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var sites []models.PlantSite
		var err error
		if flowIDStr := ctx.Query("flow_id"); flowIDStr != "" {
			flowID, parseErr := strconv.ParseInt(flowIDStr, 10, 64)
			if parseErr != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid flow_id"})
				return
			}
			sites, err = db.ListPlantSitesByFlow(ctx.Request.Context(), flowID)
		} else {
			sites, err = db.ListPlantSites(ctx.Request.Context())
		}
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, sites)
	}
}

func getPlantSiteHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		site, err := db.GetPlantSite(ctx.Request.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "plant site not found"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, site)
	}
}

func createPlantSiteHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req plantSiteRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		site, err := db.CreatePlantSite(ctx.Request.Context(), models.CreatePlantSiteParams{
			FlowID: req.FlowID,
			X:      req.X,
			Y:      req.Y,
			Z:      req.Z,
		})
		if err != nil {
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "referenced flow does not exist"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusCreated, site)
	}
}

func updatePlantSiteHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req plantSiteRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		site, err := db.UpdatePlantSite(ctx.Request.Context(), models.UpdatePlantSiteParams{
			FlowID: req.FlowID,
			X:      req.X,
			Y:      req.Y,
			Z:      req.Z,
			ID:     id,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "plant site not found"})
				return
			}
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "referenced flow does not exist"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, site)
	}
}

func deletePlantSiteHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := db.DeletePlantSite(ctx.Request.Context(), id); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.Status(http.StatusNoContent)
	}
}

func listPlantsHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var plants []models.Plant
		var err error
		if siteIDStr := ctx.Query("plant_site_id"); siteIDStr != "" {
			siteID, parseErr := strconv.ParseInt(siteIDStr, 10, 64)
			if parseErr != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid plant_site_id"})
				return
			}
			plants, err = db.ListPlantsBySite(ctx.Request.Context(), siteID)
		} else {
			plants, err = db.ListPlants(ctx.Request.Context())
		}
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, plants)
	}
}

func getPlantHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		plant, err := db.GetPlant(ctx.Request.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "plant not found"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, plant)
	}
}

func createPlantHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req plantRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		plant, err := db.CreatePlant(ctx.Request.Context(), models.CreatePlantParams{
			PlantSiteID: req.PlantSiteID,
			PlantedOn:   toNullTime(req.PlantedOn),
		})
		if err != nil {
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "referenced plant site does not exist"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusCreated, plant)
	}
}

func updatePlantHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req plantRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		plant, err := db.UpdatePlant(ctx.Request.Context(), models.UpdatePlantParams{
			PlantSiteID: req.PlantSiteID,
			PlantedOn:   toNullTime(req.PlantedOn),
			ID:          id,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "plant not found"})
				return
			}
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "referenced plant site does not exist"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, plant)
	}
}

func deletePlantHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := db.DeletePlant(ctx.Request.Context(), id); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.Status(http.StatusNoContent)
	}
}

func listEnclosureAirTemperaturesHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		limit := int64(100)
		if limitStr := ctx.Query("limit"); limitStr != "" {
			if parsed, parseErr := strconv.ParseInt(limitStr, 10, 64); parseErr == nil && parsed > 0 {
				limit = parsed
			}
		}
		readings, err := db.ListEnclosureAirTemperatures(ctx.Request.Context(), models.ListEnclosureAirTemperaturesParams{
			EnclosureID: id,
			Limit:       limit,
		})
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, readings)
	}
}

func listEnclosureAirHumidityHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		limit := int64(100)
		if limitStr := ctx.Query("limit"); limitStr != "" {
			if parsed, parseErr := strconv.ParseInt(limitStr, 10, 64); parseErr == nil && parsed > 0 {
				limit = parsed
			}
		}
		readings, err := db.ListEnclosureAirHumidity(ctx.Request.Context(), models.ListEnclosureAirHumidityParams{
			EnclosureID: id,
			Limit:       limit,
		})
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, readings)
	}
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
