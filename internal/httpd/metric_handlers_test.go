package httpd

import (
	"database/sql"
	"math"
	"testing"
	"time"
)

func TestAveragePointsOverIntervals(t *testing.T) {
	start := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(3 * time.Hour)

	at := func(offset time.Duration, value float64) metricPoint {
		return metricPoint{
			timestamp: sql.NullTime{Time: start.Add(offset), Valid: true},
			value:     value,
		}
	}

	t.Run("averages readings within each interval", func(t *testing.T) {
		// Deliberately descending to prove the result does not depend on
		// the order returned by the database.
		points := []metricPoint{
			at(150*time.Minute, 30),
			at(90*time.Minute, 20),
			at(30*time.Minute, 10),
		}

		got := averagePointsOverIntervals(points, start, end, 3)

		if len(got) != 3 {
			t.Fatalf("expected 3 intervals, got %d: %#v", len(got), got)
		}

		for i, want := range []float64{10, 20, 30} {
			if got[i].Value != want {
				t.Errorf("interval %d: expected value %v, got %v", i, want, got[i].Value)
			}
			wantTimestamp := start.Add(time.Duration(i) * time.Hour).Format(time.RFC3339)
			if got[i].Timestamp != wantTimestamp {
				t.Errorf("interval %d: expected timestamp %s, got %s", i, wantTimestamp, got[i].Timestamp)
			}
		}
	})

	t.Run("averages multiple readings in the same interval", func(t *testing.T) {
		points := []metricPoint{
			at(10*time.Minute, 10),
			at(20*time.Minute, 20),
			at(70*time.Minute, 100),
		}

		got := averagePointsOverIntervals(points, start, end, 3)
		if len(got) != 2 {
			t.Fatalf("expected 2 non-empty intervals, got %d: %#v", len(got), got)
		}

		if math.Abs(got[0].Value-15) > 1e-9 {
			t.Errorf("expected first interval average 15, got %v", got[0].Value)
		}
		if got[1].Value != 100 {
			t.Errorf("expected second interval value 100, got %v", got[1].Value)
		}
	})

	t.Run("ignores invalid and out-of-window readings", func(t *testing.T) {
		points := []metricPoint{
			{timestamp: sql.NullTime{Valid: false}, value: 1000},
			at(-time.Minute, 1000),
			at(3*time.Hour, 1000),
			at(30*time.Minute, 5),
		}

		got := averagePointsOverIntervals(points, start, end, 3)
		if len(got) != 1 {
			t.Fatalf("expected 1 interval, got %d: %#v", len(got), got)
		}
		if got[0].Value != 5 {
			t.Errorf("expected value 5, got %v", got[0].Value)
		}
	})

	t.Run("returns no intervals for an invalid window or limit", func(t *testing.T) {
		points := []metricPoint{at(30*time.Minute, 5)}

		for _, limit := range []int64{0, -1} {
			if got := averagePointsOverIntervals(points, start, end, limit); len(got) != 0 {
				t.Errorf("limit %d: expected no intervals, got %#v", limit, got)
			}
		}

		if got := averagePointsOverIntervals(points, end, start, 3); len(got) != 0 {
			t.Errorf("reversed window: expected no intervals, got %#v", got)
		}
	})
}
