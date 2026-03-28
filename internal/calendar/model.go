package calendar

import "time"

type CalendarListEntry struct {
	ID       string
	Summary  string
	Primary  bool
	Selected bool
}

type Event struct {
	ID           string
	CalendarID   string
	CalendarName string
	Summary      string
	StartTime    time.Time
	EndTime      time.Time
	AllDay       bool
	Status       string
	HTMLLink     string
}
