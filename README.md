# 🎬 Discord TMDB Bot

A full-featured Discord bot that brings [The Movie Database (TMDB)](https://www.themoviedb.org/) directly into your server. Browse trending and popular titles, search with live autocomplete, drill into rich detail cards, find trailers, see cast & crew, and check where to stream — all without leaving Discord.

> This product uses the TMDB API but is not endorsed or certified by TMDB.  
> Streaming availability data is powered by [JustWatch](https://www.justwatch.com/).

---

## ✨ Features

| Feature           | Description                                                                   |
| ----------------- | ----------------------------------------------------------------------------- |
| `/popular`        | Browse paginated lists of popular movies or TV shows                          |
| `/trending`       | See what\'s hot today or this week                                            |
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

See what\'s trending over a time window.

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
nano .env          # or open in any editor

# 3. Run
go run .
```
