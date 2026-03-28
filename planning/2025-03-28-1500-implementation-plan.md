# Implementation Plan - First Draft Execution

Date: 2025-03-28 15:00
Status: Draft for execution
Depends on: `planning/2025-03-28-1430-initial-scope.md`

## Purpose

Turn the agreed MVP into a build-ready implementation path using:
- Go
- templ
- Tailwind CSS
- htmx
- Google OAuth and Google Calendar API

Constraints carried forward:
- Multi-user support
- No database for MVP
- Stateless cookie sessions
- No refresh token persistence
- Calendar selection stays in request state, not in the cookie

## MVP Outcome

Users can:
1. Sign in via Google.
2. Choose one or more accessible calendars.
3. Query events by `start` and `end` period.
4. View events in a table.
5. Export matching events as CSV.

## Execution Strategy

Build in thin vertical slices so the app is runnable early:
1. App bootstrap, config, and routing.
2. Auth flow and session cookie.
3. Calendar discovery and selectable calendars UI.
4. Events query and Google API integration.
5. htmx table partial and CSV export.
6. Hardening and deployment artifacts.

Each phase ends with a short verification checklist.

## Proposed Project Layout

```text
.
├── cmd/
│   └── web/
│       └── main.go
├── internal/
│   ├── app/
│   │   ├── config.go
│   │   ├── server.go
│   │   └── routes.go
│   ├── auth/
│   │   ├── google.go
│   │   ├── handler.go
│   │   └── state.go
│   ├── session/
│   │   ├── cookie.go
│   │   └── middleware.go
│   ├── calendar/
│   │   ├── client.go
│   │   ├── query.go
│   │   └── model.go
│   ├── csvexport/
│   │   └── writer.go
│   └── web/
│       ├── handler.go
│       └── viewmodel.go
├── views/
│   ├── pages/
│   │   └── home.templ
│   ├── partials/
│   │   ├── events_table.templ
│   │   ├── flash.templ
│   │   ├── empty_state.templ
│   │   └── calendar_list.templ
│   └── layout.templ
├── assets/
│   ├── css/
│   │   └── input.css
│   └── dist/
│       └── app.css
├── scripts/
│   ├── dev.sh
│   └── build.sh
├── deploy/
│   ├── google-calendar-analyzer.service
│   └── caddy.example.conf
├── .env.example
├── go.mod
├── tailwind.config.js
├── package.json
└── README.md
```

Notes:
- `views/*.templ` compile to Go during build and dev.
- No persistence folder is needed for MVP.

## Config Contract

Environment variables:
- `APP_ENV` (`development` | `production`)
- `APP_ADDR` (example: `:8080`)
- `APP_BASE_URL` (example: `https://calendar.example.com`)
- `SESSION_SECRET` (32+ random bytes, base64 or raw string)
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- `GOOGLE_REDIRECT_URL` (must match Google OAuth client settings)

Optional:
- `SESSION_COOKIE_NAME` (default: `gca_session`)
- `SESSION_MAX_AGE_SECONDS` (default: `3600`)
- `MAX_QUERY_RANGE_DAYS` (default: `31`)

## Session Payload

Keep minimal and bounded:
- `sub` (Google subject)
- `email`
- `access_token`
- `access_token_expiry` (unix timestamp UTC)

Requirements:
- Signed and encrypted cookie payload.
- Flags: `Secure`, `HttpOnly`, `SameSite=Lax`, scoped path `/`.
- Reject session if the token is expired; redirect to login.
- Do not store calendar selections in the cookie.

## Route Plan

- `GET /`
  - Render the page with login state, calendar checklist, query form, and results placeholder.
- `GET /auth/google/login`
  - Build OAuth URL with `state`; redirect.
- `GET /auth/google/callback`
  - Validate `state`, exchange `code`, set session cookie, redirect `/`.
- `POST /auth/logout`
  - Clear cookie and redirect `/`.
- `POST /events/query`
  - Validate input, call Google API, return events table partial.
- `GET /events/export.csv`
  - Validate input, call Google API, stream CSV attachment.

## Query Contract

Input parameters:
- `start` (required, RFC3339 or HTML `datetime-local` converted to UTC)
- `end` (required)
- `calendar_ids[]` (required, one or more selected calendars)
- `q` (optional, accepted but ignored in MVP filtering)
- `match_mode` (optional, accepted but ignored in MVP filtering)

Validation:
- `start < end`
- Range <= `MAX_QUERY_RANGE_DAYS`
- At least one calendar selected
- Reject malformed dates with a clear error message

## Calendar Discovery Plan

Use Calendar List API after login:
- `calendarList.list`
- Show accessible calendars as checkboxes
- Pre-check the primary calendar by default if useful

Behavior:
- Calendar selection is part of the query payload.
- The cookie remains auth-only and as small as possible.
- The user can re-select calendars on each query submission.

## Google API Integration Plan

For each selected calendar:
- Call `events.list`
- `singleEvents=true`
- `orderBy=startTime`
- `timeMin=start`
- `timeMax=end`
- Paginate with `nextPageToken` until exhausted

Map response into internal model:
- `id`
- `calendar_id`
- `summary`
- `start_time`
- `end_time`
- `all_day`
- `status`
- `html_link`

Then combine results into one normalized slice for table rendering and CSV export.

## UI/UX Plan

Home page:
- Sign in/out controls
- Accessible calendar checklist
- Query form (`start`, `end`)
- `Fetch events` button using htmx POST to `/events/query`
- `Export CSV` link/button using the same query params

Partials:
- Empty state partial for no data
- Error flash partial for validation, auth, or API errors
- Events table partial for successful query
- Calendar list partial after login

Behavior:
- On token expiry, show a message and prompt re-login.
- Keep table columns stable to simplify CSV mapping.

## CSV Export Contract

Filename:
- `events-YYYYMMDDTHHMMSSZ.csv`

Columns (v1):
- `Calendar`
- `Event ID`
- `Title`
- `Start (UTC)`
- `End (UTC)`
- `All Day`
- `Status`
- `Event URL`

Rules:
- UTF-8 CSV
- RFC4180-compliant escaping
- Always include a header row
- Match table rows for the same filter inputs

## Phase-By-Phase Build Checklist

### Phase 1 - Bootstrap

Deliverables:
- `go mod init`
- HTTP server, router, health route, config loader
- templ and Tailwind wiring

Verify:
- App starts locally.
- `GET /` renders a static scaffold page.

### Phase 2 - Auth + Session

Deliverables:
- Google login and callback handlers
- OAuth `state` generation and validation
- Cookie issue, parse, and clear

Verify:
- Login succeeds and returns to `/`.
- Logout clears session.
- Invalid `state` is rejected.

### Phase 3 - Calendar Discovery

Deliverables:
- `calendarList.list` client call
- Calendar checklist UI
- Query form includes selected `calendar_ids[]`

Verify:
- User sees accessible calendars after login.
- At least one calendar can be selected and submitted.

### Phase 4 - Events Query

Deliverables:
- Query validation
- Google Calendar client wrapper
- `/events/query` returns table partial

Verify:
- Valid range returns rows.
- Empty range returns empty-state partial.
- Expired token path redirects to login prompt.

### Phase 5 - CSV Export

Deliverables:
- `/events/export.csv` endpoint
- Shared query path with table endpoint

Verify:
- CSV downloads successfully.
- CSV row count matches table result for the same filters.

### Phase 6 - Hardening + Deploy

Deliverables:
- `.env.example`
- systemd unit file
- Caddy reverse proxy sample
- README runbook

Verify:
- Build binary and run under service.
- HTTPS termination works via reverse proxy.

## Testing Plan

Automated:
- Unit tests for query parsing and validation.
- Unit tests for session cookie encode/decode and expiry checks.
- Unit tests for calendar selection handling.
- Unit tests for CSV writer escaping and headers.

Manual:
- OAuth happy path.
- Token expiry behavior.
- Calendar selection and query flow.
- Date range edge cases.
- CSV import sanity check in Google Sheets.

## Risks and Mitigations

- OAuth misconfiguration, especially redirect URI mismatch
  - Mitigation: document exact callback URL and env examples.
- Cookie too large if payload grows
  - Mitigation: keep only minimal auth fields, no profile blobs.
- Timezone confusion in filtering
  - Mitigation: convert to UTC internally and display timezone labels.
- API quota or transient failures
  - Mitigation: explicit user-facing errors and retry guidance.

## Definition of Done

- Code compiles and runs locally.
- End-to-end flow works: login -> calendar selection -> query -> table -> CSV export.
- No DB dependency is introduced.
- Expired access token triggers a clear re-auth flow.
- Deployment examples exist for VPS usage.

## Immediate Next Step

Start implementation with Phase 1 and Phase 2 in one pass:
1. Scaffold the app and config.
2. Wire OAuth and session handling.
3. Leave query and CSV stubs returning placeholder responses.

This yields a testable auth baseline before Calendar integration.
