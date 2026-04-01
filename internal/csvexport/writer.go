package csvexport

import (
	"encoding/csv"
	"io"
	"time"

	"github.com/gjtiquia/google-calendar-analyzer/internal/calendar"
)

var header = []string{
	"Date",
	"Title",
	"Start time",
	"End time",
	"Duration (hrs)",
	"Link",
}

// WriteEvents writes RFC4180 CSV with UTF-8 BOM for spreadsheet compatibility.
// loc controls how date and clock columns are formatted (client-selected IANA zone).
func WriteEvents(w io.Writer, events []calendar.Event, loc *time.Location) error {
	bom := []byte{0xEF, 0xBB, 0xBF}
	if _, err := w.Write(bom); err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	if err := cw.Write(header); err != nil {
		return err
	}
	for _, e := range events {
		row := []string{
			calendar.FormatEventDate(e, loc),
			e.Summary,
			calendar.FormatEventStartTime(e, loc),
			calendar.FormatEventEndTime(e, loc),
			calendar.FormatDurationHours(e),
			e.HTMLLink,
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
