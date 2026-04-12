# Keklik Backend

Initial Go service skeleton for the Keklik backend, structured as a modular monolith with DDD-inspired package boundaries.

## CI and image publishing

GitHub Actions is split into two workflows:

- Pull requests run `.github/workflows/ci.yml`.
- Pushes to the default branch, `master`, and version tags matching `v*` run `.github/workflows/publish.yml`.

Both workflows execute the standard Go validation steps:

- `gofmt` verification
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/api`

The publish workflow builds the root `Dockerfile` and pushes images to GitHub Container Registry as `ghcr.io/<owner>/keklik` using the repository `GITHUB_TOKEN`.

Published tags include:

- `sha-<full-commit-sha>` for every published image
- `latest` for the default branch
- the Git tag name for version tags such as `v1.2.3`

If the repository default branch changes from `master`, update the branch filter and `latest` tag condition in `.github/workflows/publish.yml`.

## Run with Docker Compose

```bash
docker compose up --build
```

This starts:

- the API on `http://localhost:8080`
- PostgreSQL on `localhost:5432`

The Compose stack uses a named Docker volume, `postgres-data`, so local database state survives container restarts.

The API container is configured with development-safe placeholder Google OAuth values so the service can boot locally before the real OAuth flow is implemented. Override them in `compose.yaml` if you need different values.

To verify the stack is healthy:

```bash
curl http://localhost:8080/healthz
```

To stop the stack:

```bash
docker compose down
```

To stop the stack and remove the local PostgreSQL volume:

```bash
docker compose down --volumes
```

## Run locally

```bash
export DATABASE_URL='postgres://keklik:keklik@localhost:5432/keklik?sslmode=disable'
export GOOGLE_OAUTH_CLIENT_ID='your-google-client-id'
export GOOGLE_OAUTH_CLIENT_SECRET='your-google-client-secret'
export GOOGLE_OAUTH_REDIRECT_URL='http://localhost:8080/auth/google/callback'
export APP_BASE_URL='http://localhost:8080'
go run ./cmd/api
```

The service listens on port `8080` by default. Override it with `HTTP_PORT`.

## Configuration

The service validates its environment contract during startup and exits immediately if required settings are missing or invalid.

| Variable | Required | Notes |
| --- | --- | --- |
| `HTTP_PORT` | No | Optional local development override. Defaults to `8080`. |
| `DATABASE_URL` | Yes | PostgreSQL DSN used by infrastructure components. |
| `GOOGLE_OAUTH_CLIENT_ID` | Yes | Google OAuth client identifier. |
| `GOOGLE_OAUTH_CLIENT_SECRET` | Yes | Google OAuth client secret. |
| `GOOGLE_OAUTH_REDIRECT_URL` | Yes | Absolute callback URL registered with Google. |
| `APP_BASE_URL` | Yes | Absolute base URL used to build application links such as family invite URLs. |

## Health check

```bash
curl http://localhost:8080/healthz
```

Expected response:

```json
{"status":"ok"}
```