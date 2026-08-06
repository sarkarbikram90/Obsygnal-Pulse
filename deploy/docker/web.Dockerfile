# Multi-stage build for React SPA Web UI
# Stage 1: Build
FROM node:22-alpine AS builder

WORKDIR /app

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ .
RUN npm run build

# Stage 2: Nginx Runtime (Unprivileged)
FROM nginxinc/nginx-unprivileged:alpine-slim

COPY --from=builder /app/dist /usr/share/nginx/html
COPY deploy/docker/nginx.conf /etc/nginx/nginx.conf

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --spider -q http://localhost:8080/healthz || exit 1

CMD ["nginx", "-g", "daemon off;"]
