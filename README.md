# go-ytdlp-bot

A Telegram bot that downloads video from URLs using [yt-dlp](https://github.com/yt-dlp/yt-dlp), with a built-in web admin dashboard.

## Features

- Automatically downloads videos from supported platforms when links are shared in Telegram chats
- Replies with error messages when downloads fail (download error, file processing, upload too large)
- H.264/AAC video encoding for universal playback (iOS/Android/Desktop)
- URL filters configurable via web dashboard (hosts, path regex, query param stripping, cookies)
- Default filters seeded on first startup for popular platforms (TikTok, YouTube, Instagram, X/Twitter, Reddit, Facebook)
- Download cache with configurable TTL to avoid re-downloading the same URL
- Access control: approve/reject Telegram groups and users, with pending approval queues for both
- Mobile-friendly web admin dashboard with:
  - Download history with pagination and filtering
  - Real-time log viewer via SSE with search
  - Live usage statistics (total downloads, success/failure ratio, top domains, daily counts)
  - Access control management (groups and users)
  - URL filter management
- SQLite database for persistence (no external DB required)
- Docker-ready with Alpine-based image

## Quick Start

### Telegram Bot Setup

1. Create a bot via [@BotFather](https://t.me/BotFather) and get the token
2. **Disable privacy mode**: send `/setprivacy` to BotFather, select your bot, choose **Disable** (required for the bot to see messages in groups)

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

3. Start:
   ```bash
   docker compose up -d
   ```

4. Open `http://localhost:8080` and log in with the dashboard credentials from your config.

5. Add the bot to a Telegram group. The group will appear as "pending" in the Access Control page for you to approve.

### Build from Source

```bash
go build -o go-ytdlp-bot ./cmd/go-ytdlp-bot
./go-ytdlp-bot -c config.yaml
```

Requires `yt-dlp` and `ffmpeg` installed on the host.

## Configuration

All configuration is in `config.yaml`. See [`config.yaml.example`](config.yaml.example) for all options.

| Section | Description |
|---------|-------------|
| `bot.token` | Telegram Bot API token |
| `storage.path` | Directory for downloaded files |
| `storage.removeAfterReply` | Delete files after sending to chat |
| `cache.ttl` | Download cache duration (e.g. `5m`) |
| `database.path` | SQLite database file path |
| `dashboard.port` | Web dashboard port (default `8080`) |
| `dashboard.username` | Dashboard login username |
| `dashboard.password` | Dashboard login password |

### URL Filters

URL filters are managed through the web dashboard at `/filters`. On first startup, default filters for popular platforms are seeded into the database. After that, all filter management happens via the dashboard.

Each filter specifies:
- **Hosts** - which domains to match (e.g. `tiktok.com`, `www.tiktok.com`)
- **Path regex** - optional regex to match URL paths (e.g. `/shorts/`)
- **Exclude query params** - strip query parameters before caching
- **Cookies file** - path to a cookies file for authenticated downloads

### Access Control

Access control is always on. Groups and users must be approved before the bot will process their requests.

- **Groups**: When the bot is added to a group or receives a message from an unapproved group, it auto-registers as "pending" in the dashboard.
- **Users**: When a user sends a video URL in a private chat, they are auto-registered as "pending".
- Approve or reject from the **Access Control** page in the dashboard.

## Dashboard

The admin dashboard is accessible at the configured port (default `8080`) and requires login. All pages are mobile-friendly.

| Page | Description |
|------|-------------|
| Home | Summary stats and quick navigation |
| Downloads | Full download history with status filtering and pagination |
| Logs | Real-time application logs with level filtering and search |
| Statistics | Live usage metrics updated via SSE |
| Access Control | Manage Telegram groups and users with pending approval queues |
| Filters | Add, edit, and delete URL filter rules |

## Project Structure

```
cmd/go-ytdlp-bot/main.go       Entry point
internal/
  bot/bot.go                    Telegram bot logic
  cache/cache.go                Download cache with TTL
  config/config.go              YAML config loading
  dashboard/                    Web dashboard (server, handlers, templates, static)
  database/                     SQLite database (migrations, access, filters, downloads)
  logger/                       Zerolog setup + DB writer for log capture
  ytdlp/                        yt-dlp integration
```

## License

See [LICENSE](LICENSE) for details.
