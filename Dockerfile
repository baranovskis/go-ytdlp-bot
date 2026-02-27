FROM golang:1.26-alpine

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Install ffmpeg
RUN apk update && apk upgrade && apk add --no-cache ffmpeg

# Copy the source code
COPY config.yaml.example ./config.yaml
COPY cmd ./cmd/
COPY internal ./internal/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-ytdlp-bot ./cmd/go-ytdlp-bot

# Run
CMD ["/go-ytdlp-bot"]