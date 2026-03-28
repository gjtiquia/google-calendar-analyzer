package calendar

import (
	"net/url"
	"testing"
	"time"
)

func TestParseQuery_ok(t *testing.T) {
	v := url.Values{}
	v.Set("start", "2026-01-01T00:00:00Z")
	v.Set("end", "2026-01-02T00:00:00Z")
	v.Add("calendar_ids", "a@group.calendar.google.com")
	v.Add("calendar_ids", "b@group.calendar.google.com")
	v.Set("q", "ignored")
	v.Set("match_mode", "ignored")

	q, err := ParseQuery(v)
	if err != nil {
		t.Fatal(err)
	}
	if !q.Start.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("start: %v", q.Start)
	}
	if len(q.CalendarIDs) != 2 {
		t.Fatalf("calendar ids: %v", q.CalendarIDs)
	}
}

func TestParseQuery_startNotBeforeEnd(t *testing.T) {
	v := url.Values{}
	v.Set("start", "2026-01-02T00:00:00Z")
	v.Set("end", "2026-01-01T00:00:00Z")
	v.Set("calendar_ids", "x")
	_, err := ParseQuery(v)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseQuery_noCalendars(t *testing.T) {
	v := url.Values{}
	v.Set("start", "2026-01-01T00:00:00Z")
	v.Set("end", "2026-01-02T00:00:00Z")
	_, err := ParseQuery(v)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSelectedIDs_dedupes(t *testing.T) {
	v := url.Values{}
	v.Add("calendar_ids", "a")
	v.Add("calendar_ids", "a")
	v.Add("calendar_ids[]", "b")
	got := SelectedIDs(v)
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("got %v", got)
	}
}
