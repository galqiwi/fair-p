FROM golang:1.23rc2-alpine3.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY ./cmd ./cmd 
COPY ./internal ./internal
COPY ./vendor ./vendor

RUN CGO_ENABLED=1 go build -race ./cmd/passer/...
RUN CGO_ENABLED=1 go build -race ./cmd/rotator/...

FROM golang:1.23rc2-alpine3.20
WORKDIR /app
COPY --from=builder /app/passer /app/rotator /app/