# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Run locally (Windows — music stub, all other features functional)
```powershell
go mod tidy
$env:CGO_ENABLED=1; go run ./cmd/
```

### Run with full music support (Linux/Docker)
```bash
docker compose up -d --build
```

### Build binary
```bash
CGO_ENABLED=1 go build -ldflags="-s -w" -o bot ./cmd/
```

There are no automated tests in this project.

## Configuration

Copy `.env.example` to `.env`. Required variables:
```env
TOKEN=          # Discord bot token
GUILD_ID=       # Discord guild ID
```

Optional:
```env
STAFF_ROLE_ID=       # Role that can see tickets
TICKET_CATEGORY_ID=  # Category channel for ticket channels
LOG_CHANNEL_ID=      # Channel to post ticket transcripts on close
DB_PATH=./data/bot.db
```

## Architecture

### Entry point and wiring
`cmd/main.go` loads config, opens the SQLite DB (which auto-migrates on startup), creates a single `commands.Handler`, wires it to discordgo via thin adapter functions in `internal/events/` (Ready, InteractionCreate, VoiceStateUpdate), adds a raw MessageCreate handler for XP gain, then blocks on SIGINT.

### Command routing
`internal/commands/handler.go` is the central hub. `NewHandler` instantiates all sub-managers and populates a `map[string]CommandFunc`. `Handle()` dispatches slash commands by name and component interactions by `CustomID`. Adding a new slash command requires: registering the handler func in `register()`, adding the `ApplicationCommand` definition in `RegisterCommands()`, and writing the handler method in the appropriate `*_cmd.go` file.

Component interactions (`bj_hit`, `bj_stand`, `close_ticket`, `transcript_ticket`) and paginated leaderboard buttons (`xplb:N`, `ecolb:N`) are dispatched in `handleComponent()`.

### Sub-systems
| Package | Responsibility |
|---|---|
| `internal/config` | Loads `.env` via godotenv |
| `internal/database` | SQLite connection (modernc pure-Go driver), WAL mode, auto-migrations on startup |
| `internal/repositories` | Raw SQL CRUD: `xp_repo.go`, `economy_repo.go`, `items_repo.go`, `bj_repo.go`, `achievements_repo.go` |
| `internal/xp` | XP gain on message (15–25 XP, 60s cooldown persisted in DB), MEE6 level formula, prestige at level 100 |
| `internal/economy` | Wallet/bank, daily (streak bonus), work (4h cooldown), atomic transfers via SQL transactions |
| `internal/games` | Coinflip, dice, slots; `blackjack.go` full engine with multi-step button interactions; `BJManager` holds in-memory sessions with a 10-min cleanup goroutine |
| `internal/achievements` | Static catalogue of achievements (`All []Achievement`) + `Manager.Check()` which unlocks idempotently; errors are logged and never bubble up to callers |
| `internal/tickets` | Opens/closes per-user ticket channels; generates TXT transcripts on close |
| `internal/music` | Per-guild `Player` + thread-safe `Queue`; `Manager` handles multi-guild state and auto-cleanup when voice channel empties; `ytdlp.go` shells out to `yt-dlp` for metadata |
| `internal/events` | Thin adapter layer — converts discordgo event callbacks into calls on `commands.Handler` interfaces |
| `internal/utils` | Shared embed builders (`EmbedFields`, color constants) and response helpers (`Respond`, `RespondEmbed`, `RespondEphemeral`) |

### Achievements
`internal/achievements/manager.go` exposes `Manager.Check(guildID, userID, key)`, which is fire-and-forget — failures are only logged. The static catalogue (`All`) defines every achievement; new achievements must be added there and triggered via `Check()` at the relevant event site (XP level-up, daily claim, blackjack outcome, etc.).

### CGO build tags for music
`internal/music/audio_cgo.go` (build tag `cgo`) calls `dgvoice.PlayAudioFile` and requires gcc + libopus. `internal/music/audio_nocgo.go` (build tag `!cgo`) is a silent stub used on Windows without gcc. Music is fully functional only in Docker (Linux image with ffmpeg + libopus).

### Database
SQLite via `modernc.org/sqlite` (pure Go, no CGO required for the DB itself). Migrations run automatically in `internal/database/migrations.go` on every startup. Schema tables: `tickets`, `guild_settings`, `user_xp` (with `last_xp_at` for persistent cooldowns), `economy`, `items`, `inventory`, `achievements`, `blackjack_sessions` (purged on every startup). All financial operations (pay, deposit, buy, sell) use SQL transactions to prevent race conditions.

### Admin permission check
Admin commands verify the Discord `Administrator` permission via the Discord API, not a hardcoded role. The check is done inside each admin command handler.
