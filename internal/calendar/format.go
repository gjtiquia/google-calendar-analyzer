package calendar

import (
	"fmt"
	"time"
)

// FormatEventDate returns the calendar date of the event start in loc (YYYY-MM-DD).
func FormatEventDate(e Event, loc *time.Location) string {
	if loc == nil {
		loc = time.UTC
	}
	return e.StartTime.In(loc).Format("2006-01-02")
}

// FormatEventStartTime returns wall time like "4:00:00 PM" in loc, or "All day" for all-day events.
func FormatEventStartTime(e Event, loc *time.Location) string {
	if e.AllDay {
		return "All day"
	}
	if loc == nil {
		loc = time.UTC
	}
	return e.StartTime.In(loc).Format("3:04:05 PM")
}

// FormatEventEndTime returns wall time like "5:00:00 PM" in loc, or "All day" for all-day events.
func FormatEventEndTime(e Event, loc *time.Location) string {
	if e.AllDay {
		return "All day"
	}
	if loc == nil {
		loc = time.UTC
	}
	return e.EndTime.In(loc).Format("3:04:05 PM")
}

// FormatDurationHours returns the event length in hours (two decimal places).
func FormatDurationHours(e Event) string {
	h := e.EndTime.Sub(e.StartTime).Hours()
	if h < 0 {
		h = 0
	}
	return fmt.Sprintf("%.2f", h)
}
