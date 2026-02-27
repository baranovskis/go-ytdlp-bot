# go-ytdlp-bot

A Telegram bot that downloads video from URLs using [yt-dlp](https://github.com/yt-dlp/yt-dlp), with a built-in web admin dashboard.

## Features

- Automatically downloads videos from supported platforms when links are shared in Telegram chats
- Replies with error messages when downloads fail (download error, file processing, upload too large)
- URL filters configurable via web dashboard (hosts, path regex, query param stripping, cookies)
- Default filters seeded on first startup for popular platforms (TikTok, YouTube, Instagram, X/Twitter, Reddit, Facebook)
- Download cache with configurable TTL to avoid re-downloading the same URL
- Web admin dashboard with:
  - Download history with pagination and filtering
  - Real-time log viewer via SSE with full context fields (URL, error reason, etc.) and search across message + fields
  - Live usage statistics (total downloads, success/failure ratio, top domains, daily counts)
  - Access control with auto-discovery: users who send video links are automatically registered as "pending" for admin approval
  - URL filter management
- SQLite database for persistence (no external DB required)
- Access control: approve/reject Telegram groups and users, with pending approval queues for both
- Docker-ready with Alpine-based image

## Quick Start

### Docker Compose (recommended)

1. Copy the example config:
   ```bash
   cp config.yaml.example config.yaml
   ```

2. Edit `config.yaml` and set your Telegram bot token:
   ```yaml
   bot:
     token: "your-bot-token-here"
   ```

3. Update `docker-compose.yml` to expose the dashboard port and persist data:
   ```yaml
   services:
     bot:
       image: ghcr.io/baranovskis/go-ytdlp-bot:main
       container_name: ytdlp-bot
       restart: unless-stopped
       security_opt:
         - no-new-privileges:true
       ports:
         - "8080:8080"
       volumes:
         - ./config.yaml:/app/config.yaml:ro
         - ./data:/app/data
   ```

4. Start:
   ```bash
   docker compose up -d
   ```

5. Open `http://localhost:8080` and log in with the dashboard credentials from your config.

### Build from Source

```bash
go build -o go-ytdlp-bot ./cmd/go-ytdlp-bot
./go-ytdlp-bot -c config.yaml
```

Requires `yt-dlp` and `ffmpeg` installed on the host (yt-dlp is auto-installed on first run).

## Configuration

All configuration is in `config.yaml`. See [`config.yaml.example`](config.yaml.example) for all options.

| Section | Description |
|---------|-------------|
| `bot.token` | Telegram Bot API token |
| `storage.path` | Directory for downloaded files |
| `storage.removeAfterReply` | Delete files after sending to chat |
| `cache.ttl` | Download cache duration (e.g. `5m`) |
| `database.path` | SQLite database file path |
| `dashboard.port` | Web dashboard port |
| `dashboard.username` | Dashboard login username |
| `dashboard.password` | Dashboard login password |
| `accessControl.enabled` | Enable group/user access control |
| `accessControl.defaultAllow` | Allow all when access control is enabled but lists are empty |

### URL Filters

URL filters are managed through the web dashboard at `/filters`. On first startup, any filters defined in `config.yaml` under `bot.filters` are seeded into the database. After that, all filter management happens via the UI.

Each filter specifies:
- **Hosts** - which domains to match (e.g. `tiktok.com`, `www.tiktok.com`)
- **Path regex** - optional regex to match URL paths (e.g. `/shorts/`)
- **Exclude query params** - strip query parameters before caching
- **Cookies file** - path to a cookies file for authenticated downloads

## Dashboard

The admin dashboard is accessible at the configured port (default `8080`) and requires login.

| Page | Description |
|------|-------------|
| Home | Summary stats and quick links |
| Downloads | Full download history with status filtering and pagination |
| Logs | Real-time application logs with level filtering and search |
| Statistics | Live usage metrics updated via SSE |
| Access Control | Manage Telegram groups and users with pending approval queues for both |
| Filters | Add, edit, and delete URL filter rules |

## Project Structure

```
cmd/go-ytdlp-bot/main.go       Entry point
internal/
  bot/bot.go                    Telegram bot logic
  cache/cache.go                Download cache with TTL
  config/config.go              YAML config loading
  dashboard/                    Web dashboard (server, handlers, templates, CSS)
  database/                     SQLite database (migrations, repos)
  logger/                       Zerolog setup + DB writer for log capture
  ytdlp/                        yt-dlp integration
```

## License

See [LICENSE](LICENSE) for details.
