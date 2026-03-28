package viewdata

import "github.com/gjtiquia/google-calendar-analyzer/internal/calendar"

// HomeView is passed to the home page template.
type HomeView struct {
	Flash           string
	OAuthConfigured bool
	LoggedIn        bool
	Email           string
	Calendars       []calendar.CalendarListEntry
	CalendarErr     string
	DefaultStart    string
	DefaultEnd      string
}
