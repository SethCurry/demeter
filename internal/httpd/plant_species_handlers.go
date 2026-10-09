package httpd

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/SethCurry/demeter/internal/models"
	"github.com/gin-gonic/gin"
)

type plantSpeciesRequest struct {
	Name         string `json:"name" binding:"required"`
	PlantGenusID *int64 `json:"plant_genus_id,omitempty"`
}

func listPlantSpeciesHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var species []models.PlantSpecy
		var err error
		if genusIDStr := ctx.Query("plant_genus_id"); genusIDStr != "" {
			genusID, parseErr := strconv.ParseInt(genusIDStr, 10, 64)
			if parseErr != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid plant_genus_id"})
				return
			}
			species, err = db.ListPlantSpeciesByGenus(ctx.Request.Context(), sql.NullInt64{Int64: genusID, Valid: true})
		} else {
			species, err = db.ListPlantSpecies(ctx.Request.Context())
		}
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, species)
	}
}

func getPlantSpeciesHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		species, err := db.GetPlantSpecies(ctx.Request.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "plant species not found"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, species)
	}
}

func createPlantSpeciesHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req plantSpeciesRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		species, err := db.CreatePlantSpecies(ctx.Request.Context(), models.CreatePlantSpeciesParams{
			Name:         req.Name,
			PlantGenusID: toNullInt64(req.PlantGenusID),
		})
		if err != nil {
			if isUniqueViolation(err) {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "plant species name already exists"})
				return
			}
			if isConstraintViolation(err) {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "referenced plant genus does not exist"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusCreated, species)
	}
}

func updatePlantSpeciesHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req plantSpeciesRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		species, err := db.UpdatePlantSpecies(ctx.Request.Context(), models.UpdatePlantSpeciesParams{
			Name:         req.Name,
			PlantGenusID: toNullInt64(req.PlantGenusID),
			ID:           id,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "plant species not found"})
				return
			}
			if isUniqueViolation(err) {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "plant species name already exists"})
				return
			}
			if isConstraintViolation(err) {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "referenced plant genus does not exist"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, species)
	}
}

func deletePlantSpeciesHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := db.DeletePlantSpecies(ctx.Request.Context(), id); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.Status(http.StatusNoContent)
	}
}
