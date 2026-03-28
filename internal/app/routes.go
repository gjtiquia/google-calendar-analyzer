package app

import (
	"net/http"

	"github.com/gjtiquia/google-calendar-analyzer/internal/auth"
	"github.com/gjtiquia/google-calendar-analyzer/internal/session"
	"github.com/gjtiquia/google-calendar-analyzer/internal/web"
)

func (s *Server) Routes() http.Handler {
	sessCfg := session.Config{
		CookieName:    s.cfg.SessionCookieName,
		SessionSecret: s.cfg.SessionSecret,
		MaxAgeSeconds: s.cfg.SessionMaxAge,
		SecureCookies: s.cfg.Env == "production",
	}
	sm := session.NewManager(sessCfg)
	authH := auth.NewHandler(
		s.cfg.GoogleClientID,
		s.cfg.GoogleClientSecret,
		s.cfg.GoogleRedirectURL,
		sm,
	)
	webH := web.NewHandler(s.cfg.OAuthConfigured(), sm)

	mux := http.NewServeMux()

	assets := http.FileServer(http.Dir("assets"))
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", assets))

	mux.HandleFunc("GET /", webH.Home)
	mux.HandleFunc("GET /privacy", webH.PrivacyPolicy)
	mux.HandleFunc("GET /terms", webH.TermsOfService)
	mux.HandleFunc("GET /healthz", webH.Healthz)

	mux.HandleFunc("GET /auth/google/login", authH.Login)
	mux.HandleFunc("GET /auth/google/callback", authH.Callback)
	mux.HandleFunc("POST /auth/logout", authH.Logout)

	mux.HandleFunc("POST /events/query", webH.EventsQuery)
	mux.HandleFunc("GET /events/export.csv", webH.ExportCSV)

	return sm.Attach(mux)
}
