package httpd

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/SethCurry/demeter/internal/models"
	"github.com/gin-gonic/gin"
)

type plantRequest struct {
	PlantSiteID int64      `json:"plant_site_id" binding:"required"`
	PlantedOn   *time.Time `json:"planted_on,omitempty"`
	SpeciesID   *int64     `json:"species_id,omitempty"`
}

func listPlantsHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var plants []models.Plant
		var err error
		siteIDStr := ctx.Query("plant_site_id")
		speciesIDStr := ctx.Query("species_id")
		switch {
		case siteIDStr != "":
			siteID, parseErr := strconv.ParseInt(siteIDStr, 10, 64)
			if parseErr != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid plant_site_id"})
				return
			}
			plants, err = db.ListPlantsBySite(ctx.Request.Context(), siteID)
		case speciesIDStr != "":
			speciesID, parseErr := strconv.ParseInt(speciesIDStr, 10, 64)
			if parseErr != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid species_id"})
				return
			}
			plants, err = db.ListPlantsBySpecies(ctx.Request.Context(), sql.NullInt64{Int64: speciesID, Valid: true})
		default:
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
			SpeciesID:   toNullInt64(req.SpeciesID),
		})
		if err != nil {
			if isConstraintViolation(err) {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "referenced plant site or species does not exist"})
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
			SpeciesID:   toNullInt64(req.SpeciesID),
			ID:          id,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "plant not found"})
				return
			}
			if isConstraintViolation(err) {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "referenced plant site or species does not exist"})
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

type plantNoteRequest struct {
	PlanID    *int64     `json:"plan_id,omitempty"`
	Content   *string    `json:"content,omitempty"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

func listPlantNotesHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var notes []models.PlantNote
		var err error
		if plantIDStr := ctx.Query("plan_id"); plantIDStr != "" {
			plantID, parseErr := strconv.ParseInt(plantIDStr, 10, 64)
			if parseErr != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid plan_id"})
				return
			}
			notes, err = db.ListPlantNotesByPlant(ctx.Request.Context(), sql.NullInt64{Int64: plantID, Valid: true})
		} else {
			notes, err = db.ListPlantNotes(ctx.Request.Context())
		}
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, notes)
	}
}

func getPlantNoteHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		note, err := db.GetPlantNote(ctx.Request.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "plant note not found"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, note)
	}
}

func createPlantNoteHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req plantNoteRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		note, err := db.CreatePlantNote(ctx.Request.Context(), models.CreatePlantNoteParams{
			PlanID:    toNullInt64(req.PlanID),
			Timestamp: toTime(req.Timestamp),
			Content:   toNullString(req.Content),
		})
		if err != nil {
			if isConstraintViolation(err) {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "referenced plant does not exist"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusCreated, note)
	}
}

func updatePlantNoteHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req plantNoteRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		note, err := db.UpdatePlantNote(ctx.Request.Context(), models.UpdatePlantNoteParams{
			PlanID:    toNullInt64(req.PlanID),
			Timestamp: toTime(req.Timestamp),
			Content:   toNullString(req.Content),
			ID:        id,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "plant note not found"})
				return
			}
			if isConstraintViolation(err) {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "referenced plant does not exist"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, note)
	}
}

func deletePlantNoteHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := db.DeletePlantNote(ctx.Request.Context(), id); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.Status(http.StatusNoContent)
	}
}
