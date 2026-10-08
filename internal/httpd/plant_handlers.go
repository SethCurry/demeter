package httpd

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/SethCurry/demeter/internal/models"
	"github.com/gin-gonic/gin"
	sqlite "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

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
