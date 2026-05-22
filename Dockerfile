FROM golang:1.26-bullseye AS builder

WORKDIR /app

# Dépendances système pour go-sqlite3 (CGO) et gopus (libopus)
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libopus-dev \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o bot ./cmd/

# ───────── Image finale ─────────
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    python3 \
    python3-pip \
    libopus0 \
    ca-certificates \
    && pip3 install --no-cache-dir yt-dlp \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/bot .

RUN mkdir -p /app/data

ENV DB_PATH=/app/data/bot.db

ENTRYPOINT ["./bot"]
