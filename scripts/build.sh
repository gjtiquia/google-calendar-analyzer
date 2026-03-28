#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

if ! command -v npm >/dev/null 2>&1; then
  echo "npm is required for Tailwind CSS builds" >&2
  exit 1
fi

npm ci
npm run css:build
go run github.com/a-h/templ/cmd/templ@latest generate
go build -o bin/google-calendar-analyzer ./cmd/web
