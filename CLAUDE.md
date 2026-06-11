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
LOG_CHANNEL_ID=      # Channel to post ticket transcripts on close and moderation logs
DB_PATH=./data/bot.db
LOG_FORMAT=text      # json or text (default: text)
LOG_LEVEL=info       # debug, info, warn, error (default: info)
HEALTH_PORT=8080     # HTTP health check port (0 = disabled)
```

Per-guild settings are stored in `guild_settings` and configured via slash commands: `levelup_channel_id`, `daily_cooldown_hours`, `work_cooldown_hours`, `max_bet`, `auto_role_id`, `roles_channel_id`, `roles_message_id`.

## Architecture

### Entry point and wiring
`cmd/main.go` loads config (which calls `logger.Init()` to set up the global slog handler), opens the SQLite DB (auto-migrates on startup), creates a single `commands.Handler`, wires it to discordgo via thin adapter functions in `internal/events/`, starts the HTTP health server, then blocks on SIGINT/SIGTERM via `signal.NotifyContext`.

On signal: health server shuts down → `dg.Close()` → `handler.Shutdown()` → `db.Close()` (defer). A 10-second context caps the whole sequence.

### Logging
All logging goes through `log/slog` (Go standard library). The global handler is initialised once in `config.Load()` via `internal/logger.Init()`, which reads `LOG_FORMAT` (json|text) and `LOG_LEVEL` (debug|info|warn|error) from env. Every log call carries a `"component"` key and relevant context attributes (`guild_id`, `user_id`, `command`, `error`). Never use `log.Printf` — use `slog.Info/Warn/Error` with key-value pairs.

### Metrics
`internal/metrics.GetMetrics()` returns the process-wide singleton (`sync.Once`). Counters use `sync/atomic.Int64`; the command map uses `sync.RWMutex`. Call sites:
- `IncrCommand(name)` — in `Handle()` on every slash command dispatch
- `IncrMessage()` — in `xp.Manager.HandleMessage()` on every message
- `IncrSpamTimeout()` — in `moderation.Monitor.Check()` on successful timeout
- `IncrDBError()` — in `xp/manager.go` and `achievements/manager.go` on DB errors

### Command routing
`internal/commands/handler.go` is the central hub. `NewHandler` instantiates all sub-managers and populates a `map[string]CommandFunc`. `Handle()` dispatches slash commands by name (and increments the metrics counter) and component interactions by `CustomID`. Adding a new slash command requires: registering the handler func in `register()`, adding the `ApplicationCommand` definition in `RegisterCommands()`, and writing the handler method in the appropriate `*_cmd.go` file.

Component interactions (`bj_hit`, `bj_stand`, `close_ticket`, `transcript_ticket`), paginated leaderboard buttons (`xplb:N`, `ecolb:N`), and role toggle buttons (`role_toggle:<roleID>`) are dispatched in `handleComponent()`.

**Leaderboard helpers**: all paginated leaderboard rendering goes through shared functions — `respondLeaderboard` (builds embed + buttons + calls `InteractionRespond`), `showXPLeaderboard` / `showEconLeaderboard` (read repo + call respondLeaderboard), `xpLeaderboardBody` (formats `[]LeaderboardEntry`), `econLeaderboardBody` (formats `[]UserEconomy`). Both `/leaderboard` and `/econleaderboard` (and their pagination buttons) call the same helpers.

### Sub-systems
| Package | Responsibility |
|---|---|
| `internal/config` | Loads `.env` via godotenv, validates TOKEN/GUILD_ID (fatal), LOG_LEVEL (warn+fallback), DB_PATH (write-check); exposes `Version = "1.0.3"` |
| `internal/logger` | Initialises the global `slog` handler once (`Init()`); reads `LOG_FORMAT` and `LOG_LEVEL` from env |
| `internal/metrics` | Process-wide singleton with atomic counters; `Snapshot()` returns top-5 commands + all counters; `FormatUptime()` formats a `time.Duration` |
| `internal/database` | SQLite connection (modernc pure-Go driver), WAL mode, auto-migrations on startup; satisfies `ChannelSettings` interface for xp.Manager |
| `internal/repositories` | Raw SQL CRUD: `xp_repo.go`, `economy_repo.go`, `items_repo.go`, `bj_repo.go`, `achievements_repo.go`, `stats_repo.go` |
| `internal/xp` | XP gain on message (15–25 XP, 60s cooldown persisted in DB), MEE6 level formula, prestige at level 100; level-up announcements sent to `levelup_channel_id` if configured |
| `internal/economy` | Wallet/bank, daily (streak bonus), work; cooldowns injected as `time.Duration` parameters (callers read from `guild_settings`); atomic transfers via SQL transactions |
| `internal/games` | Coinflip, dice, slots; `blackjack.go` full engine with multi-step button interactions; `BJManager` holds in-memory sessions with a 10-min cleanup goroutine; `enforceMaxBet` checks `guild_settings.max_bet` |
| `internal/achievements` | Static catalogue of 22 achievements (`All []Achievement`) + `Manager.Check()` which unlocks idempotently; errors are logged and never bubble up to callers |
| `internal/moderation` | Anti-spam monitor (in-memory counters, Discord timeout, `sync.Mutex` + ticker); moderation logger (ban/unban/timeout embeds → `LOG_CHANNEL_ID`); roles manager (auto-role on join, interactive role-select embed with toggle buttons) |
| `internal/tickets` | Opens/closes per-user ticket channels; generates TXT transcripts on close |
| `internal/music` | Per-guild `Player` + thread-safe `Queue`; `Manager` handles multi-guild state and auto-cleanup when voice channel empties; `ytdlp.go` shells out to `yt-dlp` for metadata |
| `internal/events` | Thin adapter layer — converts discordgo event callbacks into calls on `commands.Handler` interfaces |
| `internal/utils` | Shared embed builders (`EmbedFields`, color constants) and response helpers (`Respond`, `RespondEmbed`, `RespondEphemeral`) |

### Achievements
`internal/achievements/manager.go` exposes `Manager.Check(guildID, userID, key)`, which is fire-and-forget — failures are only logged. The static catalogue (`All`) defines 22 achievements; new achievements must be added there and triggered via `Check()` at the relevant event site. For counter-based achievements (msg_100/500/1000, bj_win_10/50), increment the counter in `StatsRepo` first and call `Check()` on the threshold value.

### CGO build tags for music
`internal/music/audio_cgo.go` (build tag `cgo`) calls `dgvoice.PlayAudioFile` and requires gcc + libopus. `internal/music/audio_nocgo.go` (build tag `!cgo`) is a silent stub used on Windows without gcc. Music is fully functional only in Docker (Linux image with ffmpeg + libopus).

### Database
SQLite via `modernc.org/sqlite` (pure Go, no CGO required for the DB itself). Migrations run automatically in `internal/database/migrations.go` on every startup. Each column addition is guarded by a `pragma_table_info` check — safe to run on any existing DB version.

Schema tables:
- `tickets`, `guild_settings` (extended with `auto_role_id`, `roles_channel_id`, `roles_message_id`, `levelup_channel_id`, `daily_cooldown_hours`, `work_cooldown_hours`, `max_bet`)
- `user_xp` (with `last_xp_at` for persistent cooldowns), `economy`, `items`, `inventory`
- `achievements` (generic key/guild/user — idempotent `INSERT OR IGNORE`)
- `blackjack_sessions` (purged on every startup)
- `role_buttons` (UNIQUE guild+role — toggle buttons for the role-select embed)
- `user_stats` (`message_count`, `bj_wins` per guild+user — drives counter achievements)

All financial operations (pay, deposit, buy, sell) use SQL transactions to prevent race conditions.

### Admin permission check
Admin commands verify the Discord `Administrator` permission via the Discord API, not a hardcoded role. The check is done inside each admin command handler. All admin money/XP commands validate `amount > 0` (or `>= 0` for setxp) before calling the manager layer.

### Security invariants
- All SQL queries use positional `?` parameters — no SQL injection possible.
- `items_repo.GetByID(guildID, id)` and `Buy` filter by `guild_id` — cross-guild item purchase is impossible.
- Admin commands validate `amount > 0` at the command handler level, before any DB call.

### Graceful shutdown
`cmd/main.go` handles SIGINT/SIGTERM via `signal.NotifyContext`. On signal, a goroutine runs the shutdown sequence within a 10-second timeout:

```
healthSrv.Shutdown(ctx)          ← 0. Fermer le health server HTTP
dg.Close()                       ← 1. Fermer la session Discord
handler.Shutdown()               ← 2. Arrêter les sub-managers
    ├── games.Manager.Shutdown()
    │       └── BJManager.Shutdown() → close(done)
    │               └── cleanup() goroutine: <-done → return
    ├── music.Manager.Shutdown()
    │       └── Player.Stop() × N active guilds
    │               └── loop() goroutine: <-p.stop → return
    └── moderation.Monitor.Shutdown() → close(done)
            └── resetLoop() goroutine: <-done → return
db.Close()                       ← 3. Fermer la DB (via defer)
```
