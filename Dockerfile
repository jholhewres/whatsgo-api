# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.24-alpine AS builder
ARG VERSION=dev
RUN apk add --no-cache gcc musl-dev git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./internal/webui/dist
RUN CGO_ENABLED=1 go build -ldflags="-s -w -X main.version=${VERSION}" -o /whatsgo ./cmd/whatsgo

# Stage 3: Runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -g '' whatsgo
COPY --from=builder /whatsgo /usr/local/bin/whatsgo
USER whatsgo
WORKDIR /home/whatsgo
EXPOSE 8550
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s \
    CMD wget -qO- http://localhost:8550/health || exit 1
ENTRYPOINT ["whatsgo"]
