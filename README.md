# 🎬 Discord TMDB Bot

A full-featured Discord bot that brings [The Movie Database (TMDB)](https://www.themoviedb.org/) directly into your server. Browse trending and popular titles, search with live autocomplete, drill into rich detail cards, find trailers, see cast & crew, and check where to stream — all without leaving Discord.

> This product uses the TMDB API but is not endorsed or certified by TMDB.  
> Streaming availability data is powered by [JustWatch](https://www.justwatch.com/).

---

## ✨ Features

| Feature           | Description                                                                   |
| ----------------- | ----------------------------------------------------------------------------- |
| `/popular`        | Browse paginated lists of popular movies or TV shows                          |
| `/trending`       | See what's hot today or this week                                             |
| `/search`         | Search with **live autocomplete** as you type                                 |
| **Detail cards**  | Rich embeds with rating, runtime, genres, budget, studios, and backdrop image |
| **🎬 Trailer**    | Opens an official YouTube trailer (ephemeral — only you see it)               |
| **👥 Cast**       | Full cast and key crew list (ephemeral)                                       |
| **📺 Watch**      | Streaming, rental and purchase options via JustWatch (US, ephemeral)          |
| **🔙 Back**       | Returns to the previous list in-place                                         |
| **Pagination**    | All list views have Previous / Next navigation                                |
| **Caching**       | In-memory TTL cache keeps repeated lookups instant                            |
| **Rate limiting** | Per-user token-bucket limits protect both Discord and TMDB                    |

---

## 🤖 Commands

### `/popular`

Browse the most popular titles right now.

| Option | Type   | Required | Values               |
| ------ | ------ | -------- | -------------------- |
| `type` | String | ✅       | `Movies`, `TV Shows` |

### `/trending`

See what's trending over a time window.

| Option   | Type   | Required | Values               |
| -------- | ------ | -------- | -------------------- |
| `type`   | String | ✅       | `Movies`, `TV Shows` |
| `window` | String | ✅       | `Today`, `This Week` |

### `/search`

Search for a specific title. Start typing — suggestions appear automatically.

| Option  | Type   | Required | Notes                                |
| ------- | ------ | -------- | ------------------------------------ |
| `type`  | String | ✅       | `Movie`, `TV Show`                   |
| `query` | String | ✅       | Live autocomplete after 2 characters |

---

## ⚡ Quick Start

```bash
# 1. Clone the repo
git clone https://github.com/D-Chamber/discord-tmdb-bot.git
cd discord-tmdb-bot

# 2. Copy the example env file and fill in your tokens
cp .env.example .env
nano .env

# 3. Run
go run .
```

That's it. The bot connects, registers commands, and logs `Slash commands registered. Bot is ready!`

---

## 📋 Prerequisites

| Requirement                                   | Version    | Notes                           |
| --------------------------------------------- | ---------- | ------------------------------- |
| [Go](https://go.dev/dl/)                      | 1.21+      | Only needed for native runs     |
| [Docker](https://docs.docker.com/get-docker/) | Any recent | Only needed for container runs  |
| Discord Bot Token                             | —          | From Discord Developer Portal   |
| TMDB API Read Access Token                    | —          | From your TMDB account settings |

---

## 🔑 Getting Your Credentials

### Discord Bot Token

1. Go to [discord.com/developers/applications](https://discord.com/developers/applications)
2. Click **New Application**, give it a name
3. Go to the **Bot** tab → **Reset Token** → copy the token
4. On the same page, scroll to **Privileged Gateway Intents** and make sure **all three are OFF**:
    - ❌ Presence Intent
    - ❌ Server Members Intent
    - ❌ Message Content Intent

**Invite the bot to your server:**

1. Go to **OAuth2 → URL Generator**
2. Check these **Scopes**: `bot` and `applications.commands`
3. Check these **Bot Permissions**: `Read Messages/View Channels`, `Send Messages`, `Embed Links`, `Read Message History`
4. Copy the generated URL and open it in your browser to add the bot to a server

> ⚠️ The `applications.commands` scope is required. Without it slash commands will not register and the bot will silently fail.

### TMDB API Read Access Token

1. Create a free account at [themoviedb.org](https://www.themoviedb.org)
2. Go to **Settings → API**
3. Copy the **API Read Access Token** — this is the long `eyJ...` string
4. ⚠️ **Do NOT use the short API Key** (the 32-char hex string). The bot uses Bearer token authentication which requires the Read Access Token.

---

## ⚙️ Configuration

Copy `.env.example` to `.env` and set the following variables:

```ini
# ── Required ──────────────────────────────────────────────────────────────────

# From Discord Developer Portal → Bot → Reset Token
DISCORD_TOKEN=your_discord_bot_token_here

# From themoviedb.org → Settings → API → API Read Access Token (the long JWT)
TMDB_ACCESS_TOKEN=your_tmdb_read_access_token_here

# ── Optional (defaults shown) ─────────────────────────────────────────────────

# How long to keep TMDB responses in the in-memory cache (minutes)
CACHE_DURATION_MINUTES=15

# Max results shown per page (Discord allows max 5 buttons per row)
MAX_RESULTS=5

# Log verbosity: info | debug | warn | error
LOG_LEVEL=info
```

> The bot also accepts `TMDB_API_KEY` for backward compatibility, but `TMDB_ACCESS_TOKEN` is preferred.

---

## 🚀 Running the Bot

### Option 1 — `go run` (development)

```bash
go run .
```

Compiles and runs in one step. Logs stream to stdout.

### Option 2 — Compiled binary

```bash
go build -o bot .
./bot
```

Single static binary, no Go toolchain needed at runtime.

### Option 3 — Docker Compose (recommended for servers)

```bash
# Build and start (detached)
docker compose up -d --build

# View logs
docker compose logs -f

# Stop
docker compose down
```

The container restarts automatically (`restart: unless-stopped`) and rotates logs at 10 MB.

### Option 4 — Docker (manual)

```bash
docker build -t discord-tmdb-bot .
docker run -d --name discord-tmdb-bot --env-file .env discord-tmdb-bot
```

---

## 🖥️ VPS Deployment with systemd

1. **Copy the binary to your server:**

```bash
scp bot user@your-server:/opt/discord-tmdb-bot/bot
scp .env user@your-server:/opt/discord-tmdb-bot/.env
```

2. **Create a systemd service** at `/etc/systemd/system/discord-tmdb-bot.service`:

```ini
[Unit]
Description=Discord TMDB Bot
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/opt/discord-tmdb-bot
ExecStart=/opt/discord-tmdb-bot/bot
Restart=always
RestartSec=5
EnvironmentFile=/opt/discord-tmdb-bot/.env

[Install]
WantedBy=multi-user.target
```

3. **Enable and start:**

```bash
sudo systemctl daemon-reload
sudo systemctl enable discord-tmdb-bot
sudo systemctl start discord-tmdb-bot
sudo journalctl -u discord-tmdb-bot -f   # tail logs
```

---

## 🔄 CI/CD with GitHub Actions

### 🚀 Build Pipeline

The included `.github/workflows/build.yml` automatically builds the bot for Linux, Windows, and macOS on every tag push (`v*`).

### 📦 Download Pre-built Binaries

If you prefer not to build from source, you can download pre-built binaries from the [Releases](https://github.com/D-Chamber/discord-tmdb-bot/releases) page. Choose the appropriate binary for your operating system:

- **Linux**: `discord-tmdb-bot-linux-amd64`
- **Windows**: `discord-tmdb-bot-windows-amd64.exe`
- **macOS**: `discord-tmdb-bot-darwin-amd64`

### 🚀 Deploy Pipeline

The included `.github/workflows/deploy.yml` automatically builds and deploys to your VPS on every push to `main`.

Add these secrets to your GitHub repository (**Settings → Secrets and variables → Actions**):

| Secret         | Description                                        |
| -------------- | -------------------------------------------------- |
| `VPS_HOST`     | IP address or hostname of your server              |
| `VPS_USERNAME` | SSH username (e.g. `ubuntu`)                       |
| `VPS_SSH_KEY`  | Contents of your private SSH key (`~/.ssh/id_rsa`) |

Push to `main` → GitHub builds the binary → SCP to server → `systemctl restart discord-tmdb-bot`.

---

## 📦 Pre-built Binaries

Pre-built binaries for Linux, Windows, and macOS are available in the [Releases](https://github.com/D-Chamber/discord-tmdb-bot/releases) section. Download the appropriate binary for your platform and run it directly.

---

## 📁 Project Structure

```
discord-tmdb-bot/
├── main.go                   # Entry point — wires everything together
├── go.mod / go.sum           # Go module dependencies
├── Dockerfile                # Multi-stage Docker build (Alpine, non-root)
├── docker-compose.yml        # Compose config with log rotation
├── .env.example              # Template for environment variables
│
├── config/
│   └── config.go             # Loads and validates env vars at startup
│
├── tmdb/
│   ├── client.go             # TMDB HTTP client — all API methods, Bearer auth
│   ├── models.go             # Go structs matching TMDB JSON responses
│   └── cache.go              # Thread-safe in-memory TTL cache
│
├── handlers/
│   ├── button.go             # Custom btn type fixing discordgo v0.27.1 emoji bug
│   ├── commands.go           # Slash command definitions, handlers, autocomplete
│   ├── components.go         # Button interaction handlers and embed builders
│   └── ratelimit.go          # Per-user token-bucket rate limiter
│
└── utils/
    └── formatters.go         # String utilities
```

---

## 🔧 Tech Stack

| Package                                                             | Purpose                                        |
| ------------------------------------------------------------------- | ---------------------------------------------- |
| [bwmarrin/discordgo](https://github.com/bwmarrin/discordgo) v0.27.1 | Discord Gateway WebSocket + REST client        |
| [joho/godotenv](https://github.com/joho/godotenv) v1.5.1            | `.env` file loading                            |
| [golang.org/x/time](https://pkg.go.dev/golang.org/x/time) v0.5.0    | Token-bucket rate limiting                     |
| Go 1.21 standard library                                            | HTTP client, JSON, sync primitives, OS signals |

No database required. All state is held in-memory and rebuilt on restart from TMDB API calls.

---

## 🗺️ How It Works

```
User types /popular
    → Discord sends InteractionCreate to bot (WebSocket)
    → Bot immediately sends DeferredChannelMessageWithSource (< 100ms)
    → Bot calls TMDB /movie/popular (checks cache first)
    → Bot edits the deferred message with embed + buttons
    → Discord shows the result

User clicks a movie button
    → Discord sends MessageComponent interaction
    → Bot sends DeferredMessageUpdate (keeps existing message)
    → Bot calls TMDB /movie/{id}?append_to_response=credits,videos,similar
    → Bot edits message with detail card + Trailer / Cast / Watch / Back buttons
```

Interaction acknowledgement always happens first — this is what prevents Discord's 3-second timeout from leaving the bot stuck thinking.

---

## 🛡️ Rate Limits

| Limiter             | Rate    | Burst | Scope          |
| ------------------- | ------- | ----- | -------------- |
| `userRateLimiter`   | 5 req/s | 10    | Slash commands |
| `buttonRateLimiter` | 2 req/s | 5     | Button presses |

Both limiters are per-user (keyed by Discord user ID) and clean up inactive entries every 10 minutes.

---

## 📝 Attribution

- Movie and TV data provided by [The Movie Database (TMDB)](https://www.themoviedb.org/)
- Streaming availability data provided by [JustWatch](https://www.justwatch.com/)

Per TMDB's attribution requirements, any public-facing use of this bot must include:

> _"This product uses the TMDB API but is not endorsed or certified by TMDB."_

---

## 📄 License

MIT — see [LICENSE](LICENSE) for details.
