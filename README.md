# Keklik Backend

Initial Go service skeleton for the Keklik backend, structured as a modular monolith with DDD-inspired package boundaries.

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