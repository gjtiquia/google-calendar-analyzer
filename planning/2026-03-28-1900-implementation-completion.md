# Implementation completion — Google Calendar Analyzer

Date: 2026-03-28  
Status: Completed (MVP aligned with `2025-03-28-1500-implementation-plan.md`)  
Related: [planning/2025-03-28-1500-implementation-plan.md](2025-03-28-1500-implementation-plan.md)

## Summary

The MVP described in the implementation plan was built: Google OAuth sign-in, stateless signed/encrypted session cookies, calendar list + multi-calendar event queries, htmx-driven table results, CSV export, deployment samples, and targeted unit tests. Local development loads an optional `.env` file via `godotenv` so `SESSION_SECRET` and Google credentials are picked up from the repo root without manual `export` for every variable.

## Stack (as planned)

- Go HTTP server, graceful shutdown (`cmd/web/main.go`)
- [templ](https://templ.guide/) templates under `views/`
- Tailwind CSS (`assets/css/input.css` → `assets/dist/app.css`)
- htmx (CDN) + small `assets/app.js` for CSV export URL building
- Google OAuth (`golang.org/x/oauth2`) and Calendar API (`google.golang.org/api/calendar/v3`)

## Packages and layout

| Area | Location |
|------|----------|
| Config | `internal/app/config.go` |
| Routing | `internal/app/routes.go` (session middleware wraps mux) |
| OAuth handlers | `internal/auth/` (`google.go`, `handler.go`, `state.go`) |
| Session cookie | `internal/session/cookie.go`, `middleware.go` |
| Calendar API + query validation | `internal/calendar/` (`client.go`, `query.go`, `model.go`, `names.go`) |
| CSV export | `internal/csvexport/writer.go` |
| HTTP handlers | `internal/web/handler.go` |
| View DTOs (no import cycles with templates) | `internal/viewdata/home.go` |
| Pages / partials | `views/layout.templ`, `views/pages/home.templ`, `views/partials/*.templ` |
| Scripts | `scripts/dev.sh`, `scripts/build.sh` |
| Deploy examples | `deploy/google-calendar-analyzer.service`, `deploy/caddy.example.conf` |

## Configuration

Environment variables match the plan’s contract, including `APP_BASE_URL`, `SESSION_SECRET`, Google OAuth trio, and optional `SESSION_COOKIE_NAME`, `SESSION_MAX_AGE_SECONDS`. See `.env.example` and README.

## Behavior notes

- **Session payload**: `sub`, `email`, `access_token`, `access_token_expiry`; cookie signed/encrypted with `github.com/gorilla/securecookie` and HKDF-derived keys from `SESSION_SECRET`. Calendar selections are **not** stored in the cookie (only in form/query state).
- **OAuth state**: Short-lived signed cookie for CSRF state on `/auth/google/*`.
- **Routes**: `GET /`, `GET /healthz`, `GET /auth/google/login`, `GET /auth/google/callback`, `POST /auth/logout`, `POST /events/query`, `GET /events/export.csv`.
- **`.env`**: On startup, `godotenv.Load()` runs from the process working directory; run the app from the repository root so `.env` is found.

## Testing

Unit tests cover query parsing/validation (`internal/calendar/query_test.go`), session encode/decode and expiry (`internal/session/cookie_test.go`), and CSV headers/escaping (`internal/csvexport/writer_test.go`). Run `go test ./...`.

## Documentation and ops

- `README.md`: runbook, Google OAuth setup steps, configuration table, deploy overview.
- `deploy/`: systemd unit template and Caddy example for HTTPS termination.

## Google Cloud (operator checklist)

- Enable **Google Calendar API** for the project.
- Configure **OAuth consent screen**; in **Testing** mode, add **Test users** under the consent screen for accounts that should be allowed to sign in.
- Create OAuth **Web** client credentials; **Authorized redirect URI** must exactly match `GOOGLE_REDIRECT_URL` (scheme, host, port, path). `localhost` vs `127.0.0.1` are different URIs.

## Follow-ups (optional, not in original MVP)

- Keyword / match-mode filtering (`q`, `match_mode` accepted but ignored per plan).
- Refresh token persistence (explicitly out of scope for MVP).
- Broader integration or E2E tests against live Google APIs.
