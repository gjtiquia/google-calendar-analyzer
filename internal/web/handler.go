package web

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/gjtiquia/google-calendar-analyzer/internal/calendar"
	"github.com/gjtiquia/google-calendar-analyzer/internal/csvexport"
	"github.com/gjtiquia/google-calendar-analyzer/internal/session"
	"github.com/gjtiquia/google-calendar-analyzer/internal/viewdata"
	"github.com/gjtiquia/google-calendar-analyzer/views/pages"
	"github.com/gjtiquia/google-calendar-analyzer/views/partials"
)

type Handler struct {
	oauthConfigured bool
	sess            *session.Manager
}

func NewHandler(oauthConfigured bool, sess *session.Manager) *Handler {
	return &Handler{
		oauthConfigured: oauthConfigured,
		sess:            sess,
	}
}

func (h *Handler) PrivacyPolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := pages.PrivacyPolicy().Render(ctx, w); err != nil {
		http.Error(w, "failed to render privacy policy", http.StatusInternalServerError)
	}
}

func (h *Handler) TermsOfService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := pages.TermsOfService().Render(ctx, w); err != nil {
		http.Error(w, "failed to render terms of service", http.StatusInternalServerError)
	}
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	p := session.PayloadFromContext(ctx)

	view := viewdata.HomeView{
		OAuthConfigured: h.oauthConfigured,
		LoggedIn:        p != nil,
		DefaultStart:    time.Now().UTC().Add(-24 * time.Hour).Format("2006-01-02"),
		DefaultEnd:      time.Now().UTC().Add(7 * 24 * time.Hour).Format("2006-01-02"),
	}
	if p != nil {
		view.Email = p.Email
		cals, err := calendar.ListCalendars(ctx, p.AccessToken)
		if err != nil {
			view.CalendarErr = friendlyAPIErr(err)
		} else {
			view.Calendars = selectDefaultCalendars(cals)
		}
	}
	if msg := r.URL.Query().Get("error"); msg != "" {
		view.Flash = flashForError(msg)
	}

	if err := pages.Home(view).Render(ctx, w); err != nil {
		http.Error(w, "failed to render home page", http.StatusInternalServerError)
		return
	}
}

func selectDefaultCalendars(cals []calendar.CalendarListEntry) []calendar.CalendarListEntry {
	out := make([]calendar.CalendarListEntry, len(cals))
	copy(out, cals)
	hasPrimary := false
	for _, c := range out {
		if c.Primary {
			hasPrimary = true
			break
		}
	}
	for i := range out {
		if out[i].Primary {
			out[i].Selected = true
		}
	}
	if !hasPrimary && len(out) > 0 {
		out[0].Selected = true
	}
	return out
}

func flashForError(code string) string {
	switch code {
	case "oauth_not_configured":
		return "Google OAuth is not configured. Set GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, and GOOGLE_REDIRECT_URL."
	case "invalid_state", "state_mismatch":
		return "Sign-in could not be verified. Please try again."
	case "token_exchange", "userinfo", "userinfo_client":
		return "Could not complete sign-in with Google. Try again."
	case "session_write":
		return "Could not save your session. Check SESSION_SECRET and try again."
	case "auth":
		return "Please sign in to export events."
	default:
		if code == "" {
			return ""
		}
		return "Something went wrong: " + code
	}
}

func friendlyAPIErr(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	if strings.Contains(s, "401") || strings.Contains(s, "invalid_grant") {
		return "Your Google session expired. Please sign in again."
	}
	return "Could not load calendars: " + s
}

func (h *Handler) EventsQuery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		h.writeEventsFragment(ctx, w, partials.Flash("Invalid form submission."))
		return
	}
	p := session.PayloadFromContext(ctx)
	if p == nil {
		h.writeEventsFragment(ctx, w, partials.Flash("Please sign in again."))
		return
	}
	if p.Expired(time.Now()) {
		h.sess.ClearSession(w)
		h.writeEventsFragment(ctx, w, partials.Flash("Your session expired. Please sign in again."))
		return
	}

	q, err := calendar.ParseQuery(r.Form)
	if err != nil {
		h.writeEventsFragment(ctx, w, partials.Flash(err.Error()))
		return
	}

	events, err := calendar.ListEventsForCalendars(ctx, p.AccessToken, q.CalendarIDs, q.Start, q.End, q.Q)
	if err != nil {
		h.writeEventsFragment(ctx, w, partials.Flash(friendlyAPIErr(err)))
		return
	}

	idToName := calendarNameLookup(r.Context(), p.AccessToken)
	events = calendar.WithCalendarNames(events, idToName)

	if len(events) == 0 {
		h.writeEventsFragment(ctx, w, partials.EmptyState())
		return
	}
	summary := calendar.SummarizeEvents(events)
	h.writeEventsFragment(ctx, w, partials.EventsResult(summary, events))
}

func calendarNameLookup(ctx context.Context, accessToken string) map[string]string {
	cals, err := calendar.ListCalendars(ctx, accessToken)
	if err != nil {
		return nil
	}
	m := make(map[string]string, len(cals))
	for _, c := range cals {
		m[c.ID] = c.Summary
	}
	return m
}

func (h *Handler) writeEventsFragment(ctx context.Context, w http.ResponseWriter, c templ.Component) {
	if err := c.Render(ctx, w); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

func (h *Handler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	p := session.PayloadFromContext(ctx)
	if p == nil {
		http.Redirect(w, r, "/?error=auth", http.StatusSeeOther)
		return
	}
	if p.Expired(time.Now()) {
		h.sess.ClearSession(w)
		http.Redirect(w, r, "/?error=auth", http.StatusSeeOther)
		return
	}

	values := r.URL.Query()
	q, err := calendar.ParseQuery(values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	events, err := calendar.ListEventsForCalendars(ctx, p.AccessToken, q.CalendarIDs, q.Start, q.End, q.Q)
	if err != nil {
		http.Error(w, friendlyAPIErr(err), http.StatusBadGateway)
		return
	}
	idToName := calendarNameLookup(ctx, p.AccessToken)
	events = calendar.WithCalendarNames(events, idToName)

	fn := fmt.Sprintf("events-%s.csv", time.Now().UTC().Format("20060102T150405Z"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fn))
	if err := csvexport.WriteEvents(w, events); err != nil {
		return
	}
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
