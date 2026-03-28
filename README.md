# Google Calendar Analyzer

Web app to sign in with Google, pick calendars, query events in a time range, view results in a table, and export CSV. Built with Go, [templ](https://templ.guide/), Tailwind CSS, and htmx. Sessions are stateless signed+encrypted cookies (no database); refresh tokens are not stored.

## Prerequisites

- Go 1.25+
- Node.js/npm (for Tailwind CSS build)
- A Google Cloud project with the **Google Calendar API** enabled
- OAuth 2.0 credentials (Web or Desktop client) with an authorized redirect URI that matches `GOOGLE_REDIRECT_URL`

## Google OAuth setup

1. In [Google Cloud Console](https://console.cloud.google.com/), enable **Google Calendar API** for your project.
2. Create **OAuth client ID** credentials (type *Web application* is typical for production behind HTTPS).
3. Add an **Authorized redirect URI** exactly equal to your app’s callback URL, for example:
   - Local: `http://localhost:8080/auth/google/callback`
   - Production: `https://your.domain/auth/google/callback`
4. Copy the client ID and client secret into your environment (see below). The same redirect URI must appear in `.env` as `GOOGLE_REDIRECT_URL` and in the Google client configuration.

## Configuration

Copy [.env.example](.env.example) to `.env` and fill in values. Important variables:

| Variable | Purpose |
|----------|---------|
| `APP_BASE_URL` | Public origin (no trailing slash); used for OAuth. Must match how users reach the app. |
| `SESSION_SECRET` | At least 32 bytes (raw or base64); used to sign and encrypt the session cookie. |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | From Google Cloud OAuth client. |
| `GOOGLE_REDIRECT_URL` | Full callback URL; must match Google Console exactly. |
| `APP_ENV` | `development` or `production` (`production` sets `Secure` on cookies). |

Optional: `SESSION_COOKIE_NAME`, `SESSION_MAX_AGE_SECONDS`.

On startup, the binary loads a **`.env` file in the current working directory** (via `godotenv`) if it exists, then reads `os.Getenv`. Run from the repo root so `.env` is found, or export variables yourself / use a process manager in production.

```bash
# install dependencies
npm install
```

## Run locally

From the repository root (so `assets/` is served correctly):

```bash
./scripts/dev.sh
```

This installs npm dependencies if needed, builds CSS, runs `templ generate`, and starts the server on `APP_ADDR` (default `:8080`).

Open `http://localhost:8080` (use the same host you put in `GOOGLE_REDIRECT_URL`).

## Build CSS / templ (when you change styles or templates)

```bash
npm run css:build
go run github.com/a-h/templ/cmd/templ@latest generate
```

## Run on a server (git pull + go run)

From the repo root (so `./assets` is found and `git rev-parse` resolves for asset cache busting):

```bash
git pull
npm ci && npm run css:build   # when frontend or templates changed
go run github.com/a-h/templ/cmd/templ@latest generate
go run ./cmd/web
```

Or build a binary: `go build -o bin/google-calendar-analyzer ./cmd/web` and run it with the same working directory as the clone.

## Deploy on a VPS (overview)

1. Clone the repo on the server and keep `assets/` beside the process working directory.
2. Create `/etc/google-calendar-analyzer.env` with the same variables as `.env.example` (production values).
3. Use [deploy/google-calendar-analyzer.service](deploy/google-calendar-analyzer.service) as a template for systemd: set `User`/`Group`, paths, and `EnvironmentFile`.
4. Put TLS in front with a reverse proxy; see [deploy/caddy.example.conf](deploy/caddy.example.conf). Set `APP_BASE_URL` and `GOOGLE_REDIRECT_URL` to your public HTTPS URL and register that redirect URI in Google Cloud.

## Health check

`GET /healthz` returns `200` with body `ok`.

## Testing

```bash
go test ./...
```

Manual checks: OAuth sign-in, token expiry (short `SESSION_MAX_AGE_SECONDS` or wait), multi-calendar query, empty range, CSV download vs table rows.
