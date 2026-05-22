# Discord Bot Inkyra — Go

Bot Discord modulaire et production-ready écrit en Go.

## Stack

- **Go 1.26** + [discordgo](https://github.com/bwmarrin/discordgo)
- **SQLite** (go-sqlite3) pour le stockage local
- **ffmpeg** + **yt-dlp** pour la musique YouTube
- **Docker** ready

## Arborescence

```
cmd/
  main.go                  ← point d'entrée
internal/
  config/config.go         ← chargement .env
  database/                ← SQLite (migrations, CRUD tickets)
  utils/                   ← helpers embeds + réponses
  events/                  ← handlers discordgo (ready, interaction, voice)
  commands/                ← toutes les slash commands
  tickets/                 ← logique métier tickets
  music/                   ← logique musique (queue, player, yt-dlp)
data/
  bot.db                   ← base SQLite (gitignorée)
```

## Configuration

Copie `.env.example` en `.env` :

```env
TOKEN=ton_token_discord
GUILD_ID=ton_guild_id

# Optionnel
STAFF_ROLE_ID=
TICKET_CATEGORY_ID=
LOG_CHANNEL_ID=
DB_PATH=./data/bot.db
```

## Prérequis (local)

- Go 1.26+
- gcc (pour go-sqlite3 via CGO)
- libopus-dev
- ffmpeg
- yt-dlp (`pip install yt-dlp`)

## Lancer en local

```bash
go mod tidy
CGO_ENABLED=1 go run ./cmd/
```

## Docker

```bash
docker compose up -d --build
```

## Commandes

| Commande             | Description                |
| -------------------- | -------------------------- |
| `/ping`              | Répond pong 🏓             |
| `/help`              | Liste les commandes        |
| `/stats`             | Stats du bot               |
| `/ticket`            | Ouvre un ticket de support |
| `/close`             | Ferme le ticket actuel     |
| `/adduser <user>`    | Ajoute un user au ticket   |
| `/removeuser <user>` | Retire un user du ticket   |
| `/play <query>`      | Joue une musique YouTube   |
| `/skip`              | Passe à la suivante        |
| `/stop`              | Arrête et vide la queue    |
| `/queue`             | Affiche la file d'attente  |
| `/pause`             | Pause la lecture           |
| `/resume`            | Reprend la lecture         |
| `/leave`             | Quitte le vocal            |

## Roadmap

- [x] Slash commands
- [x] Système de tickets
- [x] Musique YouTube
- [ ] Anti-spam
- [ ] Auto-roles
- [ ] Menus et embeds avancés
