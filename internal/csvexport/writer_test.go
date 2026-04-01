package csvexport

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/gjtiquia/google-calendar-analyzer/internal/calendar"
)

func TestWriteEvents_headerAndEscaping(t *testing.T) {
	events := []calendar.Event{
		{
			CalendarID:   "cal",
			CalendarName: "My Cal",
			ID:           "ev1",
			Summary:      `Title, with "quotes"`,
			StartTime:    time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
			EndTime:      time.Date(2026, 3, 1, 11, 0, 0, 0, time.UTC),
			AllDay:       false,
			Status:       "confirmed",
			HTMLLink:     "https://example.com/e",
		},
	}
	var buf bytes.Buffer
	if err := WriteEvents(&buf, events, time.UTC); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.HasPrefix(s, "\ufeff") {
		t.Fatal("expected utf-8 bom")
	}
	r := csv.NewReader(strings.NewReader(strings.TrimPrefix(s, "\ufeff")))
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows: %v", rows)
	}
	if rows[0][0] != "Date" || rows[1][0] != "2026-03-01" {
		t.Fatalf("header/row: %v", rows)
	}
	if rows[1][1] != `Title, with "quotes"` {
		t.Fatalf("field: %q", rows[1][1])
	}
	// Times formatted in UTC for stable assertions.
	if rows[1][4] != "1.00" {
		t.Fatalf("duration: %q", rows[1][4])
	}
	if rows[1][5] != "https://example.com/e" {
		t.Fatalf("link: %q", rows[1][5])
	}
}
