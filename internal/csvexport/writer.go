package csvexport

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/gjtiquia/google-calendar-analyzer/internal/calendar"
)

var header = []string{
	"Calendar",
	"Event ID",
	"Title",
	"Start (UTC)",
	"End (UTC)",
	"All Day",
	"Status",
	"Event URL",
}

// WriteEvents writes RFC4180 CSV with UTF-8 BOM for spreadsheet compatibility.
func WriteEvents(w io.Writer, events []calendar.Event) error {
	bom := []byte{0xEF, 0xBB, 0xBF}
	if _, err := w.Write(bom); err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	if err := cw.Write(header); err != nil {
		return err
	}
	for _, e := range events {
		cal := e.CalendarName
		if cal == "" {
			cal = e.CalendarID
		}
		row := []string{
			cal,
			e.ID,
			e.Summary,
			formatTime(e.StartTime, e.AllDay),
			formatTime(e.EndTime, e.AllDay),
			fmt.Sprintf("%t", e.AllDay),
			e.Status,
			e.HTMLLink,
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func formatTime(t time.Time, allDay bool) string {
	if allDay {
		return t.UTC().Format("2006-01-02")
	}
	return t.UTC().Format(time.RFC3339)
}
