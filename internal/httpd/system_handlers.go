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

type systemRequest struct {
	Name        string `json:"name" binding:"required"`
	EnclosureID int64  `json:"enclosure_id" binding:"required"`
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

// systemReadingLimit parses the optional "limit" query parameter, applying
// a sane default and floor so sensor reading lists stay bounded.
func systemReadingLimit(ctx *gin.Context) int64 {
	limit := int64(100)
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if parsed, parseErr := strconv.ParseInt(limitStr, 10, 64); parseErr == nil && parsed > 0 {
			limit = parsed
		}
	}
	return limit
}

func listSystemPHHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		readings, err := db.ListSystemPH(ctx.Request.Context(), models.ListSystemPHParams{
			SystemID: id,
			Limit:    systemReadingLimit(ctx),
		})
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, readings)
	}
}

func listSystemECHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		readings, err := db.ListSystemEC(ctx.Request.Context(), models.ListSystemECParams{
			SystemID: id,
			Limit:    systemReadingLimit(ctx),
		})
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, readings)
	}
}

func listSystemWaterTemperatureHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		readings, err := db.ListSystemWaterTemperature(ctx.Request.Context(), models.ListSystemWaterTemperatureParams{
			SystemID: id,
			Limit:    systemReadingLimit(ctx),
		})
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, readings)
	}
}
