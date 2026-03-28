package calendar

import "fmt"

// SummaryRow is one labeled value in the fetch summary table. Add rows by
// extending EventSummary and Rows().
type SummaryRow struct {
	Label string
	Value string
}

// EventSummary aggregates stats for a set of fetched events. Add fields here
// as you compute more metrics; update Rows() to expose them in stable order.
type EventSummary struct {
	EventCount int
	TotalHours float64
}

// SummarizeEvents computes aggregate stats for the given events.
func SummarizeEvents(events []Event) EventSummary {
	var s EventSummary
	s.EventCount = len(events)
	for _, e := range events {
		h := e.EndTime.Sub(e.StartTime).Hours()
		if h < 0 {
			h = 0
		}
		s.TotalHours += h
	}
	return s
}

// Rows returns display rows for the summary table (order is user-facing).
func (s EventSummary) Rows() []SummaryRow {
	return []SummaryRow{
		{Label: "Events", Value: fmt.Sprintf("%d", s.EventCount)},
		{Label: "Total hours", Value: fmt.Sprintf("%.2f", s.TotalHours)},
	}
}
