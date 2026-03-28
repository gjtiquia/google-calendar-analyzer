package partials

import (
	"fmt"

	"github.com/gjtiquia/google-calendar-analyzer/internal/calendar"
)

// FormatEventDate returns the calendar date of the event start in UTC (YYYY-MM-DD).
func FormatEventDate(e calendar.Event) string {
	return e.StartTime.UTC().Format("2006-01-02")
}

// FormatDurationHours returns the event length in hours (two decimal places).
func FormatDurationHours(e calendar.Event) string {
	h := e.EndTime.Sub(e.StartTime).Hours()
	if h < 0 {
		h = 0
	}
	return fmt.Sprintf("%.2f", h)
}
