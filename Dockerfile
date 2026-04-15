FROM golang:1.22-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/keklik-api ./cmd/api

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && adduser -D -g '' appuser

COPY --from=builder /out/keklik-api /usr/local/bin/keklik-api

USER appuser

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/keklik-api"]