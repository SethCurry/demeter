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
