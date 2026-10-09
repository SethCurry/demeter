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

type flowRequest struct {
	Name         string `json:"name" binding:"required"`
	SystemID     int64  `json:"system_id" binding:"required"`
	ParentFlowID *int64 `json:"parent_flow_id,omitempty"`
}

func listFlowPlantsHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		plants, err := db.ListFlowPlants(ctx.Request.Context(), id)
		ctx.JSON(http.StatusOK, plants)
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
