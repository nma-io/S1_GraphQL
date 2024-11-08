package main

import (
	"log"
	"time"
)

// Helper Functions that are used throughout the rest of the program

func getEpochDaysAgo(days int) int64 {
	return time.Now().AddDate(0, 0, -days).Unix() * 1000
}

func parseDetectedAt(detectedAtStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, detectedAtStr)
}

func calculateStartTimestamp(minutesAgo int) int64 {
	return time.Now().Add(-time.Duration(minutesAgo) * time.Minute).UnixMilli()
}

func parseDateToEpoch(dateStr string) int64 {
	layout := "2006-01-02"
	t, err := time.Parse(layout, dateStr)
	if err != nil {
		log.Fatalf("Invalid date format: %v. Use YYYY-MM-DD.", err)
	}
	return t.Unix() * 1000 // Convert to milliseconds
}
