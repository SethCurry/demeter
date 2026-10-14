package httpd

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/SethCurry/demeter/internal/models"
	"github.com/gin-gonic/gin"
)

type MetricRequest struct {
	StartTime time.Time
	EndTime time.Time
	Limit int
}

type MetricResult struct {
	Timestamp string
	Value float64
}

func parseQueryTime(ctx *gin.Context, paramName string, ifEmpty string) (time.Time, error) {
	timeAsStr := ctx.Query(paramName)
	if timeAsStr == "" {
		timeAsStr = ifEmpty
	}

	parsed, err := time.Parse(time.RFC3339, timeAsStr)
	return parsed, err
}


func metricEnclosureHumidity(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var limit int64 = 24

		limitStr := ctx.Query("limit")
		if limitStr != "" {
			limit, err = strconv.ParseInt(limitStr, 10, 64)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
				return
			}
		}


		startTime, err := parseQueryTime(ctx, "start", time.Now().Add(time.Hour *-24).Format(time.RFC3339))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		endTime, err := parseQueryTime(ctx, "end", time.Now().Format(time.RFC3339))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		humidities, err := db.MetricsEnclosureAirHumidities(ctx, models.MetricsEnclosureAirHumiditiesParams{
			EnclosureID: id,
			Timestamp: sql.NullTime{
				Valid: true,
				Time: startTime,
			},
			Timestamp_2: sql.NullTime{
				Valid: true,
				Time: endTime,
			},
		})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		totalDuration := endTime.Sub(startTime)
		totalSeconds := int(totalDuration.Seconds())

		intervalSeconds := totalSeconds/int(limit)

		var values []MetricResult
		intervalStartsAt := startTime
		intervalEndsAt := startTime.Add(time.Second*time.Duration(intervalSeconds))
		var currentSum float64
		var numValues int

		for _, v := range humidities {
			if numValues > 0 && v.Timestamp.Valid && v.Timestamp.Time.After(intervalEndsAt) {
				values = append(values, MetricResult{
					Timestamp: intervalStartsAt.Format(time.RFC3339),
					Value: currentSum/float64(numValues),
				})
				intervalStartsAt = intervalEndsAt
				intervalEndsAt = intervalStartsAt.Add(time.Second*time.Duration(intervalSeconds))
				currentSum = 0
				numValues = 0
			} else {
				currentSum += v.HumidityRh
				numValues++
			}
		}
		ctx.JSON(http.StatusOK, values)
	}
}


func metricEnclosureAirTemperature(db *models.Queries) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var limit int64 = 24

		limitStr := ctx.Query("limit")
		if limitStr != "" {
			limit, err = strconv.ParseInt(limitStr, 10, 64)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
				return
			}
		}


		startTime, err := parseQueryTime(ctx, "start", time.Now().Add(time.Hour *-24).Format(time.RFC3339))
		if err != nil {
			ctx.AbortWithError(500, err)
		}

		endTime, err := parseQueryTime(ctx, "end", time.Now().Format(time.RFC3339))
		if err != nil {
			ctx.AbortWithError(500, err)
		}

		temps, err := db.MetricsEnclosureAirTemperatures(ctx, models.MetricsEnclosureAirTemperaturesParams{
			EnclosureID: id,
			Timestamp: sql.NullTime{
				Valid: true,
				Time: startTime,
			},
			Timestamp_2: sql.NullTime{
				Valid: true,
				Time: endTime,
			},
		})
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}

		totalDuration := endTime.Sub(startTime)
		totalSeconds := int(totalDuration.Seconds())

		intervalSeconds := totalSeconds/int(limit)

		var values []MetricResult
		intervalStartsAt := startTime
		intervalEndsAt := startTime.Add(time.Second*time.Duration(intervalSeconds))
		var currentSum float64
		var numValues int

		for _, v := range temps {
			if numValues > 0 && v.Timestamp.Valid && v.Timestamp.Time.After(intervalEndsAt) {
				values = append(values, MetricResult{
					Timestamp: intervalStartsAt.Format(time.RFC3339),
					Value: currentSum/float64(numValues),
				})
				intervalStartsAt = intervalEndsAt
				intervalEndsAt = intervalStartsAt.Add(time.Second*time.Duration(intervalSeconds))
				currentSum = 0
				numValues = 0
			} else {
				currentSum += v.TemperatureC
				numValues++
			}
		}
		ctx.JSON(http.StatusOK, values)
	}
}
