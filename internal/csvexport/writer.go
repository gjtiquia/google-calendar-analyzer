package csvexport

import (
	"encoding/csv"
	"io"

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
		row := []string{
			calendar.FormatEventDate(e),
			e.Summary,
			calendar.FormatEventStartTime(e),
			calendar.FormatEventEndTime(e),
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
