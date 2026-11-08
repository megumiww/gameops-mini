# GameOps Mini

GameOps Mini is a small Go backend for managing region-aware game announcements and live events for overseas players.

## Day 1: HTTP Service Skeleton

The first version starts an HTTP server with a health check endpoint:

```text
GET /healthz
```

Run it locally:

```bash
go run ./cmd/server
```

Then visit `http://localhost:8080/healthz` or run:

```bash
curl http://localhost:8080/healthz
```

Expected response:

```json
{"status":"ok","service":"gameops-mini"}
```

## Day 2: Game API

The service now stores games in a lightweight JSON-backed local store, so the project can run without external services during development.

```text
GET  /api/games
GET  /api/games/{id}
POST /api/games
```

Create a game on Windows PowerShell:

```powershell
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/games `
  -ContentType 'application/json' `
  -Body '{"name":"Skyforge","region":"US"}'
```

## Day 3: Announcements and Delivery

The project now includes region-aware announcements and a small administrator API.

```text
GET    /api/games/{id}/announcements
POST   /api/admin/announcements
DELETE /api/admin/announcements/{id}
```

Administrator requests use the `GAMEOPS_ADMIN_TOKEN` environment variable. The local default is `dev-token`; set a new value before deployment.

```powershell
$env:GAMEOPS_ADMIN_TOKEN = 'replace-this-token'
go run ./cmd/server
```

The repository is intentionally small: one Go service, local file-backed persistence, validation, tests, and a Dockerfile. This keeps the project easy to understand while demonstrating API design, persistence, authorization, and delivery basics.

## Delivery Checklist

- Health check: `GET /healthz`
- Game CRUD foundation
- Region-aware announcement API
- Bearer-token administrator access
- Input validation and HTTP error handling
- Automated tests and `go vet`
- Docker build configuration

## Engineering Notes

- Announcement listing supports region and language filters plus bounded pagination.
- Every response receives an `X-Request-ID` for tracing a request through logs.
- The server logs method, path, duration, and request sequence information.
- Shutdown is graceful: the server stops accepting new work and allows active requests five seconds to finish.
- The API contract is documented in [`openapi.yaml`](./openapi.yaml).
