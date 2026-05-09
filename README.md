# Keklik

Baby sleep tracker. Go backend + React/TypeScript frontend.

## Run the frontend locally

Requires Node 20+. If you use nvm: `nvm use 22`.

```bash
cd frontend
npm install
npm run dev
```

The dev server starts at `http://localhost:5173`.

---

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

The API container is configured with development-safe placeholder Google OAuth values. Override them in `compose.yaml` if you need real OAuth credentials.

When `ENABLE_SWAGGER_UI=true` (the default in Compose), the API docs are available at `http://localhost:8080/swagger`.

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

Requires [air](https://github.com/air-verse/air) for hot reload.

```bash
cd backend
cp .env.example .env
# fill in real values if needed, then:
air
```

The service listens on port `8080` by default. Override it with `HTTP_PORT` in `.env`.

## Configuration

The service validates its environment contract during startup and exits immediately if required settings are missing or invalid.

| Variable | Required | Notes |
| --- | --- | --- |
| `HTTP_PORT` | No | Defaults to `8080`. |
| `DATABASE_URL` | Yes | PostgreSQL DSN. |
| `JWT_SIGNING_KEY` | Yes | Secret used to sign auth tokens. |
| `GOOGLE_OAUTH_CLIENT_ID` | Yes | Google OAuth client identifier. |
| `GOOGLE_OAUTH_CLIENT_SECRET` | Yes | Google OAuth client secret. |
| `GOOGLE_OAUTH_REDIRECT_URL` | Yes | Absolute callback URL registered with Google. |
| `APP_BASE_URL` | Yes | Absolute base URL used to build application links such as invite URLs. |
| `FRONTEND_URL` | Yes | Absolute URL of the frontend, used for CORS and post-auth redirects. |
| `ENABLE_TEST_AUTH` | No | Enables the test-only `POST /auth/test/login` endpoint. Defaults to `false`. |
| `ENABLE_SWAGGER_UI` | No | Serves API docs at `/swagger`. Defaults to `false`. |
| `FAMILY_INVITE_LINK_EXPIRY` | No | How long invite links are valid (e.g. `48h`). Defaults to `168h` (7 days). |

## Test-only auth flow

For local development and QA, you can enable a non-Google auth path that mints a normal application session without completing the real OAuth flow.

The endpoint is disabled by default and is intended only for local or explicitly approved environments.

1. Start the service with `ENABLE_TEST_AUTH=true`.
1. Request a test session:

```bash
curl -X POST http://localhost:8080/auth/test/login \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"qa-user"}'
```

Expected response:

```json
{"token":"<bearer-token>","account_id":"<account-id>"}
```

Use the returned token exactly like a normal authenticated session:

```bash
curl http://localhost:8080/sleep-sessions \
  -H 'Authorization: Bearer <bearer-token>'
```

The test login provisions or reuses an account whose subject ID is derived as `test:<identifier>`. If you need the test user to line up with seeded local data, seed related records with the same subject ID value.

## Health check

```bash
curl http://localhost:8080/healthz
```

Expected response:

```json
{"status":"ok"}
```
