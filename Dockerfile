# syntax=docker/dockerfile:1
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /src
RUN apk add --no-cache git

COPY go.mod ./
ENV GOPRIVATE=github.com/raspberry-vision/*
ENV GONOSUMDB=github.com/raspberry-vision/*
ENV GOPROXY=direct
RUN --mount=type=ssh \
    mkdir -p -m 0700 /root/.ssh \
    && ssh-keyscan github.com >> /root/.ssh/known_hosts \
    && git config --global url."git@github.com:".insteadOf "https://github.com/" \
    && go mod download

COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/api ./cmd/api

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

COPY --from=builder /bin/api /app/api
COPY web/templates /app/web/templates
COPY web/static /app/web/static
COPY docs /app/docs
COPY CHANGELOG.md /app/CHANGELOG.md

RUN adduser -D -g '' appuser && \
    mkdir -p /data && \
    chown -R appuser:appuser /app /data

USER appuser

ENV PORT=8080 \
    DATABASE_PATH=/data/payments.db \
    BASE_URL=http://localhost:8080 \
    STATIC_DIR=/app/web/static \
    TEMPLATE_DIR=/app/web/templates \
    DOCS_DIR=/app/docs \
    CHANGELOG_PATH=/app/CHANGELOG.md

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:$${PORT}/health || exit 1

ENTRYPOINT ["/app/api"]
