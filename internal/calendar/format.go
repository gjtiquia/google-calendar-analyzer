package calendar

import (
	"fmt"
	"time"
)

// FormatEventDate returns the calendar date of the event start in UTC (YYYY-MM-DD).
func FormatEventDate(e Event) string {
	return e.StartTime.UTC().Format("2006-01-02")
}

// FormatEventStartTime returns local wall time like "4:00:00 PM", or "All day" for all-day events.
func FormatEventStartTime(e Event) string {
	if e.AllDay {
		return "All day"
	}
	return e.StartTime.In(time.Local).Format("3:04:05 PM")
}

// FormatEventEndTime returns local wall time like "5:00:00 PM", or "All day" for all-day events.
func FormatEventEndTime(e Event) string {
	if e.AllDay {
		return "All day"
	}
	return e.EndTime.In(time.Local).Format("3:04:05 PM")
}

// FormatDurationHours returns the event length in hours (two decimal places).
func FormatDurationHours(e Event) string {
	h := e.EndTime.Sub(e.StartTime).Hours()
	if h < 0 {
		h = 0
	}
	return fmt.Sprintf("%.2f", h)
}
