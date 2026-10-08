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
