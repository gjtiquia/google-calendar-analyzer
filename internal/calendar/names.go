package calendar

// WithCalendarNames copies events and fills CalendarName from idToName when present.
func WithCalendarNames(events []Event, idToName map[string]string) []Event {
	if len(idToName) == 0 {
		return events
	}
	out := make([]Event, len(events))
	copy(out, events)
	for i := range out {
		if n, ok := idToName[out[i].CalendarID]; ok && n != "" {
			out[i].CalendarName = n
		}
	}
	return out
}
