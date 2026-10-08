package httpd

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// "Simple" notes: enclosure_note, system_note, flow_note.
//
// These three tables share the same shape: an auto-incrementing id, a nullable
// timestamp (schema default CURRENT_TIMESTAMP), and a nullable content text
// column. Notably they have NO foreign key to their nominal parent
// (enclosure/system/flow) -- see the note at the bottom of this file.
//
// To avoid repeating near-identical handlers, the CRUD handlers below are
// generic and parameterized by the model type (and params type for
// create/update), with the specific sqlc method passed in.
// ---------------------------------------------------------------------------

type simpleNoteRequest struct {
	Content   *string    `json:"content,omitempty"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

func listSimpleNotesHandler[T any](name string, fn func(context.Context) ([]T, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		items, err := fn(ctx.Request.Context())
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, items)
	}
}

func getSimpleNoteHandler[T any](name string, fn func(context.Context, int64) (T, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		item, err := fn(ctx.Request.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": name + " not found"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, item)
	}
}

func createSimpleNoteHandler[T any, P any](name string, build func(simpleNoteRequest) P, fn func(context.Context, P) (T, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req simpleNoteRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		item, err := fn(ctx.Request.Context(), build(req))
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusCreated, item)
	}
}

func updateSimpleNoteHandler[T any, P any](name string, build func(simpleNoteRequest, int64) P, fn func(context.Context, P) (T, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req simpleNoteRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		item, err := fn(ctx.Request.Context(), build(req, id))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": name + " not found"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, item)
	}
}

func deleteSimpleNoteHandler(name string, fn func(context.Context, int64) error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := fn(ctx.Request.Context(), id); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.Status(http.StatusNoContent)
	}
}

// ---------------------------------------------------------------------------
// plant_note: nullable plan_id FK (to plant), NOT NULL timestamp, nullable content.
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Schema note: enclosure_note, system_note, and flow_note have no foreign key
// column linking them to their respective parent (enclosure/system/flow).
// They are therefore standalone rows identifiable only by id. The endpoints
// below expose CRUD against the tables as they exist; if the schema is later
// amended to add a parent FK, the list handlers should gain a query filter and
// the create/update request structs a required parent id.
// ---------------------------------------------------------------------------
