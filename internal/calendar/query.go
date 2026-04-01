package calendar

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type Query struct {
	Start       time.Time
	End         time.Time
	CalendarIDs []string
	// Q is passed to the Google Calendar API events.list "q" parameter (free-text search). Empty omits it.
	Q string
	// Location is the IANA timezone used to format event dates/times for display (from client "tz"). Never nil.
	Location *time.Location
}

// ParseQuery parses start, end, calendar IDs, and optional search string from form values (POST body or GET query).
func ParseQuery(values url.Values) (Query, error) {
	startRaw := strings.TrimSpace(values.Get("start"))
	endRaw := strings.TrimSpace(values.Get("end"))
	if startRaw == "" || endRaw == "" {
		return Query{}, errors.New("start and end are required")
	}
	start, err := parseDateTime(startRaw)
	if err != nil {
		return Query{}, fmt.Errorf("invalid start: %w", err)
	}
	end, err := parseDateTime(endRaw)
	if err != nil {
		return Query{}, fmt.Errorf("invalid end: %w", err)
	}
	if !start.Before(end) {
		return Query{}, errors.New("start must be before end")
	}

	var ids []string
	ids = append(ids, values["calendar_ids"]...)
	ids = append(ids, values["calendar_ids[]"]...)
	seen := map[string]struct{}{}
	var out []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if len(out) == 0 {
		return Query{}, errors.New("select at least one calendar")
	}

	q := strings.TrimSpace(values.Get("q"))

	loc := loadDisplayLocation(values.Get("tz"))

	return Query{
		Start:       start.UTC(),
		End:         end.UTC(),
		CalendarIDs: out,
		Q:           q,
		Location:    loc,
	}, nil
}

// loadDisplayLocation parses an IANA timezone name for formatting. Empty or invalid falls back to UTC.
func loadDisplayLocation(tz string) *time.Location {
	tz = strings.TrimSpace(tz)
	if tz == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC
	}
	return loc
}

func parseDateTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC(), nil
	}
	const localLayout = "2006-01-02T15:04"
	if t, err := time.ParseInLocation(localLayout, s, time.UTC); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.UTC); err == nil {
		return t.UTC(), nil
	}
	return time.Time{}, errors.New("use RFC3339 or datetime-local (UTC)")
}

// SelectedIDs returns unique, non-empty calendar_ids from form values.
func SelectedIDs(values url.Values) []string {
	var raw []string
	raw = append(raw, values["calendar_ids"]...)
	raw = append(raw, values["calendar_ids[]"]...)
	seen := map[string]struct{}{}
	var out []string
	for _, id := range raw {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
