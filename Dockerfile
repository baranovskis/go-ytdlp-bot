# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd/
COPY internal ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -o /go-ytdlp-bot ./cmd/go-ytdlp-bot

# Runtime: CPU-only (default)
FROM alpine:latest AS cpu

RUN apk update && apk upgrade && apk add --no-cache ffmpeg

WORKDIR /app

COPY --from=builder /go-ytdlp-bot /go-ytdlp-bot
COPY config.yaml.example ./config.yaml

CMD ["/go-ytdlp-bot"]

# Runtime: VAAPI (Intel/AMD GPU)
FROM alpine:latest AS vaapi

RUN apk update && apk upgrade && apk add --no-cache \
    ffmpeg \
    libva \
    intel-media-driver \
    mesa-va-gallium

WORKDIR /app

COPY --from=builder /go-ytdlp-bot /go-ytdlp-bot
COPY config.yaml.example ./config.yaml

CMD ["/go-ytdlp-bot"]
