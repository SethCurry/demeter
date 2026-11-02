package httpd

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/SethCurry/demeter/internal/models"
	"github.com/gin-gonic/gin"
)

type MetricRequest struct {
	StartTime time.Time
	EndTime   time.Time
	Limit     int
}

type MetricResult struct {
	Timestamp string
	Value     float64
}

// metricPoint is a single timestamped reading of an enclosure metric.
type metricPoint struct {
	timestamp sql.NullTime
	value     float64
}

func parseQueryTime(ctx *gin.Context, paramName string, ifEmpty string) (time.Time, error) {
	timeAsStr := ctx.Query(paramName)
	if timeAsStr == "" {
		timeAsStr = ifEmpty
	}

	parsed, err := time.Parse(time.RFC3339, timeAsStr)
	return parsed, err
}

func parseQueryLimit(ctx *gin.Context) (int64, error) {
	const (
		defaultLimit int64 = 24
		maxLimit     int64 = 10000
	)

	limitStr := ctx.Query("limit")
	if limitStr == "" {
		return defaultLimit, nil
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		return 0, err
	}

	if limit < 1 || limit > maxLimit {
		return 0, fmt.Errorf("limit must be between 1 and %d", maxLimit)
	}

	return limit, nil
}

// metricByID returns a handler that serves a metric averaged over
// equal-length intervals for the entity identified by the "id" route
// parameter (an enclosure or a system). fetch retrieves the raw readings
// for that entity within the requested window.
func metricByID(fetch func(ctx context.Context, id int64, start, end time.Time) ([]metricPoint, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		limit, err := parseQueryLimit(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}

		now := time.Now()
		startTime, err := parseQueryTime(ctx, "start", now.Add(-24*time.Hour).Format(time.RFC3339))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		endTime, err := parseQueryTime(ctx, "end", now.Format(time.RFC3339))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !startTime.Before(endTime) {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "start must be before end"})
			return
		}

		points, err := fetch(ctx, id, startTime, endTime)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, averagePointsOverIntervals(points, startTime, endTime, limit))
	}
}

// averagePointsOverIntervals splits the window [start, end) into limit
// equal-length intervals and returns the average of the readings in each
// interval that has at least one reading. Intervals are emitted in
// chronological order regardless of the order of the input points.
func averagePointsOverIntervals(points []metricPoint, start, end time.Time, limit int64) []MetricResult {
	values := []MetricResult{}
	if limit < 1 || !start.Before(end) {
		return values
	}

	interval := end.Sub(start) / time.Duration(limit)
	if interval <= 0 {
		interval = time.Nanosecond
	}

	type bucket struct {
		sum   float64
		count int
	}

	buckets := make([]bucket, limit)
	for _, point := range points {
		if !point.timestamp.Valid {
			continue
		}

		timestamp := point.timestamp.Time
		if timestamp.Before(start) || !timestamp.Before(end) {
			continue
		}

		index := int(timestamp.Sub(start) / interval)
		if index < 0 || index >= len(buckets) {
			// The final interval absorbs any remainder when the window does
			// not divide evenly into limit intervals.
			index = len(buckets) - 1
		}

		buckets[index].sum += point.value
		buckets[index].count++
	}

	for index, bucket := range buckets {
		if bucket.count == 0 {
			continue
		}

		values = append(values, MetricResult{
			Timestamp: start.Add(time.Duration(index) * interval).Format(time.RFC3339),
			Value:     bucket.sum / float64(bucket.count),
		})
	}

	return values
}

func metricEnclosureHumidity(db *models.Queries) gin.HandlerFunc {
	return metricByID(func(ctx context.Context, id int64, start, end time.Time) ([]metricPoint, error) {
		humidities, err := db.MetricsEnclosureAirHumidities(ctx, models.MetricsEnclosureAirHumiditiesParams{
			EnclosureID: id,
			Timestamp: sql.NullTime{
				Valid: true,
				Time:  start,
			},
			Timestamp_2: sql.NullTime{
				Valid: true,
				Time:  end,
			},
		})
		if err != nil {
			return nil, err
		}

		points := make([]metricPoint, 0, len(humidities))
		for _, humidity := range humidities {
			points = append(points, metricPoint{timestamp: humidity.Timestamp, value: humidity.HumidityRh})
		}

		return points, nil
	})
}

func metricEnclosureAirTemperature(db *models.Queries) gin.HandlerFunc {
	return metricByID(func(ctx context.Context, id int64, start, end time.Time) ([]metricPoint, error) {
		temps, err := db.MetricsEnclosureAirTemperatures(ctx, models.MetricsEnclosureAirTemperaturesParams{
			EnclosureID: id,
			Timestamp: sql.NullTime{
				Valid: true,
				Time:  start,
			},
			Timestamp_2: sql.NullTime{
				Valid: true,
				Time:  end,
			},
		})
		if err != nil {
			return nil, err
		}

		points := make([]metricPoint, 0, len(temps))
		for _, temp := range temps {
			points = append(points, metricPoint{timestamp: temp.Timestamp, value: temp.TemperatureC})
		}

		return points, nil
	})
}

func metricSystemPH(db *models.Queries) gin.HandlerFunc {
	return metricByID(func(ctx context.Context, id int64, start, end time.Time) ([]metricPoint, error) {
		readings, err := db.MetricsSystemPH(ctx, models.MetricsSystemPHParams{
			SystemID: id,
			Timestamp: sql.NullTime{
				Valid: true,
				Time:  start,
			},
			Timestamp_2: sql.NullTime{
				Valid: true,
				Time:  end,
			},
		})
		if err != nil {
			return nil, err
		}

		points := make([]metricPoint, 0, len(readings))
		for _, reading := range readings {
			points = append(points, metricPoint{timestamp: reading.Timestamp, value: reading.Ph})
		}

		return points, nil
	})
}

func metricSystemEC(db *models.Queries) gin.HandlerFunc {
	return metricByID(func(ctx context.Context, id int64, start, end time.Time) ([]metricPoint, error) {
		readings, err := db.MetricsSystemEC(ctx, models.MetricsSystemECParams{
			SystemID: id,
			Timestamp: sql.NullTime{
				Valid: true,
				Time:  start,
			},
			Timestamp_2: sql.NullTime{
				Valid: true,
				Time:  end,
			},
		})
		if err != nil {
			return nil, err
		}

		points := make([]metricPoint, 0, len(readings))
		for _, reading := range readings {
			points = append(points, metricPoint{timestamp: reading.Timestamp, value: reading.Ec})
		}

		return points, nil
	})
}

func metricSystemWaterTemperature(db *models.Queries) gin.HandlerFunc {
	return metricByID(func(ctx context.Context, id int64, start, end time.Time) ([]metricPoint, error) {
		temps, err := db.MetricsSystemWaterTemperature(ctx, models.MetricsSystemWaterTemperatureParams{
			SystemID: id,
			Timestamp: sql.NullTime{
				Valid: true,
				Time:  start,
			},
			Timestamp_2: sql.NullTime{
				Valid: true,
				Time:  end,
			},
		})
		if err != nil {
			return nil, err
		}

		points := make([]metricPoint, 0, len(temps))
		for _, temp := range temps {
			points = append(points, metricPoint{timestamp: temp.Timestamp, value: temp.WaterTemperatureC})
		}

		return points, nil
	})
}

func metricPlantSiteLux(db *models.Queries) gin.HandlerFunc {
	return metricByID(func(ctx context.Context, id int64, start, end time.Time) ([]metricPoint, error) {
		readings, err := db.MetricsPlantSiteLux(ctx, models.MetricsPlantSiteLuxParams{
			PlantSiteID: id,
			Timestamp: sql.NullTime{
				Valid: true,
				Time:  start,
			},
			Timestamp_2: sql.NullTime{
				Valid: true,
				Time:  end,
			},
		})
		if err != nil {
			return nil, err
		}

		points := make([]metricPoint, 0, len(readings))
		for _, reading := range readings {
			points = append(points, metricPoint{timestamp: reading.Timestamp, value: float64(reading.Lux)})
		}

		return points, nil
	})
}
