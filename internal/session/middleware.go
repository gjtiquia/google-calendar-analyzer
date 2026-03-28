package session

import (
	"context"
	"net/http"
	"time"
)

type ctxKey int

const payloadKey ctxKey = 1

func PayloadFromContext(ctx context.Context) *Payload {
	p, _ := ctx.Value(payloadKey).(*Payload)
	return p
}

func WithPayload(ctx context.Context, p *Payload) context.Context {
	return context.WithValue(ctx, payloadKey, p)
}

// Attach reads the session cookie, clears it if expired, and stores payload in context (may be nil).
func (m *Manager) Attach(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, err := m.ReadSession(r)
		if err != nil {
			m.ClearSession(w)
			next.ServeHTTP(w, r.WithContext(WithPayload(r.Context(), nil)))
			return
		}
		if p != nil && p.Expired(time.Now()) {
			m.ClearSession(w)
			next.ServeHTTP(w, r.WithContext(WithPayload(r.Context(), nil)))
			return
		}
		next.ServeHTTP(w, r.WithContext(WithPayload(r.Context(), p)))
	})
}
