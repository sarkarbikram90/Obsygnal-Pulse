# Multi-stage build for GPU Simulator
# Stage 1: Build
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -o /app/simulator ./cmd/simulator/

# Stage 2: Runtime
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -u 1000 appuser

WORKDIR /app
COPY --from=builder /app/simulator .

USER appuser

ENTRYPOINT ["./simulator"]
