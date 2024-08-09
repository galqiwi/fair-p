FROM golang:1.23rc2-alpine3.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY ./cmd ./cmd 
COPY ./internal ./internal
# COPY ./vendor ./vendor

RUN CGO_ENABLED=0 go build ./cmd/passer/...
RUN CGO_ENABLED=0 go build ./cmd/rotator/...

FROM python:3.9.13-alpine3.15
WORKDIR /app
COPY --from=builder /app/passer /app/rotator /app/