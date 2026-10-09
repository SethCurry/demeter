package httpd

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/SethCurry/demeter/internal/models"
	"github.com/gin-gonic/gin"
)

type plantGenusRequest struct {
	Name string `json:"name" binding:"required"`
}

func listPlantGeneraHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		genera, err := db.ListPlantGenera(ctx.Request.Context())
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, genera)
	}
}

func getPlantGenusHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		genus, err := db.GetPlantGenus(ctx.Request.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "plant genus not found"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, genus)
	}
}

func createPlantGenusHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req plantGenusRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		genus, err := db.CreatePlantGenus(ctx.Request.Context(), req.Name)
		if err != nil {
			if isUniqueViolation(err) {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "plant genus name already exists"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusCreated, genus)
	}
}

func updatePlantGenusHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req plantGenusRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		genus, err := db.UpdatePlantGenus(ctx.Request.Context(), models.UpdatePlantGenusParams{
			Name: req.Name,
			ID:   id,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "plant genus not found"})
				return
			}
			if isUniqueViolation(err) {
				ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "plant genus name already exists"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, genus)
	}
}

func deletePlantGenusHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := db.DeletePlantGenus(ctx.Request.Context(), id); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.Status(http.StatusNoContent)
	}
}
