# Keklik Backend

Initial Go service skeleton for the Keklik backend, structured as a modular monolith with DDD-inspired package boundaries.

## Run locally

```bash
go run ./cmd/api
```

The service listens on port `8080` by default. Override it with `HTTP_PORT`.

## Health check

```bash
curl http://localhost:8080/healthz
```

Expected response:

```json
{"status":"ok"}
```