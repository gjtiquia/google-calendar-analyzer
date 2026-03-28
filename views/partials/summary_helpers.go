package partials

import "github.com/gjtiquia/google-calendar-analyzer/internal/calendar"

// summaryRows bridges EventSummary into templ templates (templ cannot call methods on types from other packages in range).
func summaryRows(s calendar.EventSummary) []calendar.SummaryRow {
	return s.Rows()
}
