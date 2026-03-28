package partials

import (
	"time"

	"github.com/gjtiquia/google-calendar-analyzer/internal/calendar"
)

func FormatStart(e calendar.Event) string {
	if e.AllDay {
		return e.StartTime.UTC().Format("2006-01-02")
	}
	return e.StartTime.UTC().Format(time.RFC3339)
}

func FormatEnd(e calendar.Event) string {
	if e.AllDay {
		return e.EndTime.UTC().Format("2006-01-02")
	}
	return e.EndTime.UTC().Format(time.RFC3339)
}

func CalendarLabel(e calendar.Event) string {
	if e.CalendarName != "" {
		return e.CalendarName
	}
	return e.CalendarID
}

func AllDayLabel(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
