package httpd

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/SethCurry/demeter/internal/models"
	"github.com/gin-gonic/gin"
	sqlite "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// toNullString converts an optional string pointer into a sql.NullString.
// A nil pointer (omitted or JSON null) becomes a NULL database value.
func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// toTime converts an optional time pointer into a non-null time.Time,
// defaulting to the current time when omitted. Used for note tables whose
// timestamp column is NOT NULL with no schema default.
func toTime(t *time.Time) time.Time {
	if t == nil {
		return time.Now()
	}
	return *t
}

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

// reqCtx is an alias to keep the generic handler signatures readable.
// ---------------------------------------------------------------------------
// plant_site_note: required plant_site_id FK, NOT NULL timestamp, nullable content.
// ---------------------------------------------------------------------------

type plantSiteNoteRequest struct {
	PlantSiteID int64      `json:"plant_site_id" binding:"required"`
	Content     *string    `json:"content,omitempty"`
	Timestamp   *time.Time `json:"timestamp,omitempty"`
}

func listPlantSiteNotesHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var notes []models.PlantSiteNote
		var err error
		if siteIDStr := ctx.Query("plant_site_id"); siteIDStr != "" {
			siteID, parseErr := strconv.ParseInt(siteIDStr, 10, 64)
			if parseErr != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid plant_site_id"})
				return
			}
			notes, err = db.ListPlantSiteNotesBySite(ctx.Request.Context(), siteID)
		} else {
			notes, err = db.ListPlantSiteNotes(ctx.Request.Context())
		}
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, notes)
	}
}

func getPlantSiteNoteHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		note, err := db.GetPlantSiteNote(ctx.Request.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "plant site note not found"})
				return
			}
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, note)
	}
}

func createPlantSiteNoteHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req plantSiteNoteRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		note, err := db.CreatePlantSiteNote(ctx.Request.Context(), models.CreatePlantSiteNoteParams{
			PlantSiteID: req.PlantSiteID,
			Timestamp:   toTime(req.Timestamp),
			Content:     toNullString(req.Content),
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
		ctx.JSON(http.StatusCreated, note)
	}
}

func updatePlantSiteNoteHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req plantSiteNoteRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		note, err := db.UpdatePlantSiteNote(ctx.Request.Context(), models.UpdatePlantSiteNoteParams{
			PlantSiteID: req.PlantSiteID,
			Timestamp:   toTime(req.Timestamp),
			Content:     toNullString(req.Content),
			ID:          id,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "plant site note not found"})
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
		ctx.JSON(http.StatusOK, note)
	}
}

func deletePlantSiteNoteHandler(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := db.DeletePlantSiteNote(ctx.Request.Context(), id); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.Status(http.StatusNoContent)
	}
}

// ---------------------------------------------------------------------------
// plant_note: nullable plan_id FK (to plant), NOT NULL timestamp, nullable content.
// ---------------------------------------------------------------------------

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
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT {
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
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT {
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

// ---------------------------------------------------------------------------
// Schema note: enclosure_note, system_note, and flow_note have no foreign key
// column linking them to their respective parent (enclosure/system/flow).
// They are therefore standalone rows identifiable only by id. The endpoints
// below expose CRUD against the tables as they exist; if the schema is later
// amended to add a parent FK, the list handlers should gain a query filter and
// the create/update request structs a required parent id.
// ---------------------------------------------------------------------------