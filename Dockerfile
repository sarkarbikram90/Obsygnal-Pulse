# Unified Multi-Target Dockerfile for Obsygnal-Pulse Platform

# -------------------------------------------------------------
# Base Go Builder
# -------------------------------------------------------------
FROM golang:1.25-alpine AS go-builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev

# -------------------------------------------------------------
# Target: api-server
# -------------------------------------------------------------
FROM go-builder AS build-api-server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -o /app/bin/api-server ./cmd/api-server/

FROM alpine:3.19 AS api-server
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -u 1000 appuser
WORKDIR /app
COPY --from=build-api-server /app/bin/api-server ./service
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --spider -q http://localhost:8080/api/v1/health || exit 1
ENTRYPOINT ["./service"]

# -------------------------------------------------------------
# Target: worker
# -------------------------------------------------------------
FROM go-builder AS build-worker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -o /app/bin/worker ./cmd/worker/

FROM alpine:3.19 AS worker
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -u 1000 appuser
WORKDIR /app
COPY --from=build-worker /app/bin/worker ./service
USER appuser
ENTRYPOINT ["./service"]

# -------------------------------------------------------------
# Target: collector
# -------------------------------------------------------------
FROM go-builder AS build-collector
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -o /app/bin/collector ./cmd/collector/

FROM alpine:3.19 AS collector
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build-collector /app/bin/collector ./service
ENTRYPOINT ["./service"]

# -------------------------------------------------------------
# Target: simulator
# -------------------------------------------------------------
FROM go-builder AS build-simulator
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -o /app/bin/simulator ./cmd/simulator/

FROM alpine:3.19 AS simulator
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -u 1000 appuser
WORKDIR /app
COPY --from=build-simulator /app/bin/simulator ./service
USER appuser
ENTRYPOINT ["./service"]

# -------------------------------------------------------------
# Target: web
# -------------------------------------------------------------
FROM node:22-alpine AS web-builder
WORKDIR /app
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM nginxinc/nginx-unprivileged:alpine-slim AS web
COPY --from=web-builder /app/dist /usr/share/nginx/html
COPY deploy/docker/nginx.conf /etc/nginx/nginx.conf
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --spider -q http://localhost:8080/healthz || exit 1
CMD ["nginx", "-g", "daemon off;"]
