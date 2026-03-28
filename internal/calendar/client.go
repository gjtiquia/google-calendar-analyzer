package calendar

import (
	"context"
	"strings"
	"time"

	"golang.org/x/oauth2"
	calapi "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// listEventTypes is the set of event types documented for events.list. Passing
// these explicitly omits any other event types the API may return (for example
// calendar "task" items that are not regular events).
// See: https://developers.google.com/workspace/calendar/api/v3/reference/events/list
var listEventTypes = []string{
	"default",
	"birthday",
	"focusTime",
	"fromGmail",
	"outOfOffice",
	"workingLocation",
}

func newService(ctx context.Context, accessToken string) (*calapi.Service, error) {
	cli := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}))
	return calapi.NewService(ctx, option.WithHTTPClient(cli))
}

// ListCalendars returns accessible calendars (calendarList.list).
func ListCalendars(ctx context.Context, accessToken string) ([]CalendarListEntry, error) {
	svc, err := newService(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	var out []CalendarListEntry
	tok := ""
	for {
		call := svc.CalendarList.List().MaxResults(250)
		if tok != "" {
			call = call.PageToken(tok)
		}
		list, err := call.Do()
		if err != nil {
			return nil, err
		}
		for _, it := range list.Items {
			if it == nil {
				continue
			}
			out = append(out, CalendarListEntry{
				ID:      it.Id,
				Summary: it.Summary,
				Primary: it.Primary,
			})
		}
		if list.NextPageToken == "" {
			break
		}
		tok = list.NextPageToken
	}
	return out, nil
}

// ListEventsForCalendars runs events.list per calendar with pagination.
// If search is non-empty, it is sent as the API's free-text q parameter.
func ListEventsForCalendars(ctx context.Context, accessToken string, calendarIDs []string, start, end time.Time, search string) ([]Event, error) {
	svc, err := newService(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	timeMin := start.Format(time.RFC3339)
	timeMax := end.Format(time.RFC3339)

	var combined []Event
	for _, calID := range calendarIDs {
		tok := ""
		for {
			call := svc.Events.List(calID).
				SingleEvents(true).
				OrderBy("startTime").
				TimeMin(timeMin).
				TimeMax(timeMax).
				MaxResults(2500).
				EventTypes(listEventTypes...)
			if search != "" {
				call = call.Q(search)
			}
			if tok != "" {
				call = call.PageToken(tok)
			}
			evs, err := call.Do()
			if err != nil {
				return nil, err
			}
			for _, e := range evs.Items {
				if e == nil {
					continue
				}
				if isExcludedNonEventItem(e) {
					continue
				}
				ev, err := mapEvent(calID, e)
				if err != nil {
					continue
				}
				combined = append(combined, ev)
			}
			if evs.NextPageToken == "" {
				break
			}
			tok = evs.NextPageToken
		}
	}
	return combined, nil
}

// isExcludedNonEventItem drops Google Calendar items that are not ordinary
// meetings/events when the API still returns them (e.g. as eventType "task"
// or with a Tasks source URL).
func isExcludedNonEventItem(e *calapi.Event) bool {
	if strings.EqualFold(e.EventType, "task") {
		return true
	}
	if e.Source != nil {
		u := strings.ToLower(e.Source.Url)
		if strings.Contains(u, "tasks.google.com") {
			return true
		}
	}
	return false
}

func mapEvent(calID string, e *calapi.Event) (Event, error) {
	st, en, allDay, err := eventTimes(e)
	if err != nil {
		return Event{}, err
	}
	summary := ""
	if e.Summary != "" {
		summary = e.Summary
	}
	status := ""
	if e.Status != "" {
		status = e.Status
	}
	link := ""
	if e.HtmlLink != "" {
		link = e.HtmlLink
	}
	return Event{
		ID:           e.Id,
		CalendarID:   calID,
		CalendarName: "",
		Summary:      summary,
		StartTime:    st,
		EndTime:      en,
		AllDay:       allDay,
		Status:       status,
		HTMLLink:     link,
	}, nil
}

func eventTimes(e *calapi.Event) (start, end time.Time, allDay bool, err error) {
	if e.Start == nil || e.End == nil {
		return time.Time{}, time.Time{}, false, errInvalidEventTime
	}
	if e.Start.Date != "" && e.End.Date != "" {
		st, err1 := time.ParseInLocation("2006-01-02", e.Start.Date, time.UTC)
		if err1 != nil {
			return time.Time{}, time.Time{}, false, err1
		}
		en, err2 := time.ParseInLocation("2006-01-02", e.End.Date, time.UTC)
		if err2 != nil {
			return time.Time{}, time.Time{}, false, err2
		}
		return st, en, true, nil
	}
	if e.Start.DateTime == "" || e.End.DateTime == "" {
		return time.Time{}, time.Time{}, false, errInvalidEventTime
	}
	st, err1 := time.Parse(time.RFC3339, e.Start.DateTime)
	if err1 != nil {
		return time.Time{}, time.Time{}, false, err1
	}
	en, err2 := time.Parse(time.RFC3339, e.End.DateTime)
	if err2 != nil {
		return time.Time{}, time.Time{}, false, err2
	}
	return st.UTC(), en.UTC(), false, nil
}

var errInvalidEventTime = errInvalidEvent{}

type errInvalidEvent struct{}

func (errInvalidEvent) Error() string { return "invalid event time" }
