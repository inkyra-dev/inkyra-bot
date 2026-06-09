# Discord Bot Inkyra — Go

Bot Discord modulaire et production-ready écrit en Go.

## Stack

- **Go 1.22+** + [discordgo](https://github.com/bwmarrin/discordgo)
- **SQLite** via `modernc.org/sqlite` (pure Go — aucun CGO requis pour la DB)
- **ffmpeg** + **yt-dlp** pour la musique YouTube (Linux/Docker uniquement)
- **Docker** ready

## Arborescence

```
cmd/
  main.go                      ← point d'entrée, wiring
internal/
  config/                      ← chargement .env
  database/                    ← connexion SQLite, migrations, CRUD tickets
  repositories/                ← CRUD SQL : xp, economy, items, bj, achievements
  xp/                          ← formule MEE6, level-up, prestige, progress bar
  economy/                     ← daily, work, dépôt/retrait, shop, transferts
  games/                       ← coinflip, dice, slots, blackjack (sessions mémoire)
  achievements/                ← catalogue statique + manager (fire-and-forget)
  tickets/                     ← ouverture/fermeture canaux, transcripts TXT
  music/                       ← player par guild, queue, yt-dlp, stub Windows
  events/                      ← adaptateurs discordgo (ready, interaction, voice)
  commands/                    ← toutes les slash commands + routing central
  utils/                       ← helpers embeds + réponses Discord
data/
  bot.db                       ← base SQLite (gitignorée)
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

## Prérequis

### Local Windows (musique stub, tout le reste fonctionnel)
- Go 1.22+
- Aucune dépendance native requise (SQLite est pure Go)

### Docker / Linux (musique complète)
- Docker + Docker Compose
- ffmpeg, libopus, yt-dlp (gérés par le Dockerfile)

## Lancer en local

```powershell
go mod tidy
$env:CGO_ENABLED=1; go run ./cmd/
```

## Docker

```bash
docker compose up -d --build
```

## Commandes

### Général
| Commande  | Description          |
|-----------|----------------------|
| `/ping`   | Répond pong 🏓       |
| `/help`   | Liste les commandes  |
| `/stats`  | Stats du bot         |

### Tickets
| Commande              | Description                    |
|-----------------------|--------------------------------|
| `/ticket`             | Ouvre un ticket de support     |
| `/close`              | Ferme le ticket actuel         |
| `/adduser <user>`     | Ajoute un utilisateur          |
| `/removeuser <user>`  | Retire un utilisateur          |

### XP
| Commande                    | Description                              |
|-----------------------------|------------------------------------------|
| `/rank [user]`              | Profil XP (niveau, rang, barre)          |
| `/leaderboard [page]`       | Classement XP paginé                     |
| `/prestige`                 | Prestige au niveau 100 (reset XP)        |

### Économie
| Commande               | Description                              |
|------------------------|------------------------------------------|
| `/balance [user]`      | Affiche portefeuille + banque            |
| `/daily`               | Récompense quotidienne (streak bonus)    |
| `/work`                | Travailler (cooldown 4h)                 |
| `/deposit <montant>`   | Déposer en banque (`all` supporté)       |
| `/withdraw <montant>`  | Retirer de la banque (`all` supporté)    |
| `/pay <user> <montant>`| Payer un autre utilisateur               |
| `/shop`                | Affiche la boutique                      |
| `/buy <id>`            | Acheter un item                          |
| `/sell <id>`           | Revendre un item                         |
| `/inventory`           | Affiche l'inventaire                     |
| `/econleaderboard`     | Classement économique paginé             |

### Mini-jeux
| Commande                | Description                              |
|-------------------------|------------------------------------------|
| `/coinflip <choix> <mise>` | Pile ou face                          |
| `/dice <mise>`          | Lance un dé (×0 à ×3)                   |
| `/slots <mise>`         | Machines à sous (7 symboles)             |
| `/blackjack <mise>`     | Blackjack interactif (Hit / Stand)       |

### Achievements
| Commande              | Description                              |
|-----------------------|------------------------------------------|
| `/achievements [user]`| Affiche les succès débloqués             |

### Admin
| Commande                   | Description                          |
|----------------------------|--------------------------------------|
| `/givemoney <user> <n>`    | Donne des coins                      |
| `/removemoney <user> <n>`  | Retire des coins                     |
| `/resetuser <user>`        | Remet XP + économie à zéro           |
| `/addxp <user> <n>`        | Ajoute de l'XP                       |
| `/removexp <user> <n>`     | Retire de l'XP                       |
| `/setlevel <user> <n>`     | Définit le niveau (0–100)            |
| `/setxp <user> <n>`        | Définit l'XP total                   |
| `/economy-reset <user>`    | Remet l'économie à zéro              |

### Musique _(Linux/Docker uniquement)_
| Commande        | Description                  |
|-----------------|------------------------------|
| `/play <query>` | Joue une musique YouTube     |
| `/skip`         | Passe à la suivante          |
| `/stop`         | Arrête et vide la queue      |
| `/queue`        | Affiche la file d'attente    |
| `/pause`        | Pause la lecture             |
| `/resume`       | Reprend la lecture           |
| `/leave`        | Quitte le vocal              |

## Roadmap

Voir [ROADMAP.md](ROADMAP.md)
