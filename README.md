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
  main.go                      ← point d'entrée, wiring, health server HTTP
internal/
  config/                      ← chargement .env, validation, constante Version
  logger/                      ← init slog global (JSON/text, niveau configurable)
  metrics/                     ← compteurs atomiques Go pur (singleton)
  database/                    ← connexion SQLite, migrations, CRUD tickets + settings
  repositories/                ← CRUD SQL : xp, economy, items, bj, achievements, stats
  xp/                          ← formule MEE6, level-up, prestige, progress bar
  economy/                     ← daily, work, dépôt/retrait, shop, transferts
  games/                       ← coinflip, dice, slots, blackjack (sessions mémoire)
  achievements/                ← catalogue statique (22 achievements) + manager fire-and-forget
  moderation/                  ← anti-spam, auto-rôles, logs de modération
  tickets/                     ← ouverture/fermeture canaux, transcripts TXT
  music/                       ← player par guild, queue, yt-dlp, stub Windows
  events/                      ← adaptateurs discordgo (ready, interaction, voice, member, bans)
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

LOG_FORMAT=text   # json ou text (défaut: text)
LOG_LEVEL=info    # debug, info, warn, error (défaut: info)
HEALTH_PORT=8080  # port health check HTTP (0 = désactivé)
```

Les paramètres suivants sont configurables **par serveur** via les commandes admin (stockés dans `guild_settings`) :

| Paramètre | Commande | Défaut |
|---|---|---|
| Canal level-up | `/setlevelupchannel` | Canal du message |
| Cooldown `/daily` | `/setdailycooldown` | 24h |
| Cooldown `/work` | `/setworkcooldown` | 4h |
| Mise maximale jeux | `/setmaxbet` | Illimité |
| Rôle automatique | `/setautorole` | — |
| Embed rôles | `/setuproles` | — |

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

## Health Check

Quand `HEALTH_PORT` est défini (défaut 8080) :

| Route | Réponse |
|---|---|
| `GET /health` | `200` JSON `{"status":"ok","uptime":"2j 4h 32m","guilds":1,"version":"1.0.3"}` |
| `GET /ready` | `200` si la session Discord est ouverte, `503` sinon |

## Commandes

### Général
| Commande  | Description |
|-----------|-------------|
| `/ping`   | Répond pong 🏓 |
| `/help`   | Liste les commandes |
| `/stats`  | Stats du bot (uptime, messages traités, top commandes, erreurs DB) |

### Tickets
| Commande              | Description |
|-----------------------|-------------|
| `/ticket`             | Ouvre un ticket de support |
| `/close`              | Ferme le ticket actuel |
| `/adduser <user>`     | Ajoute un utilisateur |
| `/removeuser <user>`  | Retire un utilisateur |

### XP
| Commande                        | Description |
|---------------------------------|-------------|
| `/rank [user]`                  | Profil XP (niveau, rang, barre) |
| `/leaderboard [type] [page]`    | Classement XP ou Économie paginé |
| `/prestige`                     | Prestige au niveau 100 (reset XP) |

### Économie
| Commande               | Description |
|------------------------|-------------|
| `/balance [user]`      | Affiche portefeuille + banque |
| `/daily`               | Récompense quotidienne (streak bonus) |
| `/work`                | Travailler (cooldown configurable, 4h) |
| `/deposit <montant>`   | Déposer en banque (`all` supporté) |
| `/withdraw <montant>`  | Retirer de la banque (`all` supporté) |
| `/pay <user> <montant>`| Payer un autre utilisateur |
| `/shop`                | Affiche la boutique |
| `/buy <id>`            | Acheter un item |
| `/sell <id>`           | Revendre un item |
| `/inventory`           | Affiche l'inventaire |
| `/econleaderboard`     | Alias classement économique paginé |

### Mini-jeux
| Commande                        | Description |
|---------------------------------|-------------|
| `/coinflip <choix> <mise>`      | Pile ou face |
| `/dice <mise>`                  | Lance un dé (×0 à ×3) |
| `/slots <mise>`                 | Machines à sous (7 symboles) |
| `/blackjack <mise>`             | Blackjack interactif (Hit / Stand) |

> La mise est vérifiée contre le plafond `/setmaxbet` si configuré.

### Achievements
| Commande               | Description |
|------------------------|-------------|
| `/achievements [user]` | Affiche les 22 succès (débloqués/verrouillés) |

### Modération _(admin)_
| Commande                                    | Description |
|---------------------------------------------|-------------|
| `/setautorole <role>`                       | Rôle attribué automatiquement à l'arrivée |
| `/setuproles`                               | Poste l'embed de sélection de rôles (épinglé) |
| `/addrolebutton <role> <label> [emoji]`     | Ajoute un bouton toggle à l'embed rôles |

> Les bans/débans et les timeouts anti-spam sont loggués dans `LOG_CHANNEL_ID`.

### Configuration _(admin)_
| Commande                      | Description |
|-------------------------------|-------------|
| `/setlevelupchannel <canal>`  | Canal dédié pour les annonces level-up & prestige |
| `/setdailycooldown <heures>`  | Cooldown `/daily` (1–168h, défaut 24h) |
| `/setworkcooldown <heures>`   | Cooldown `/work` (1–168h, défaut 4h) |
| `/setmaxbet <montant>`        | Mise maximale pour les jeux (0 = illimité) |
| `/config`                     | Affiche tous les réglages du serveur (éphémère) |

### Admin
| Commande                   | Description |
|----------------------------|-------------|
| `/givemoney <user> <n>`    | Donne des coins |
| `/removemoney <user> <n>`  | Retire des coins |
| `/resetuser <user>`        | Remet XP + économie à zéro |
| `/addxp <user> <n>`        | Ajoute de l'XP |
| `/removexp <user> <n>`     | Retire de l'XP |
| `/setlevel <user> <n>`     | Définit le niveau (0–100) |
| `/setxp <user> <n>`        | Définit l'XP total |
| `/economy-reset <user>`    | Remet l'économie à zéro |

### Musique _(Linux/Docker uniquement)_
| Commande        | Description |
|-----------------|-------------|
| `/play <query>` | Joue une musique YouTube |
| `/skip`         | Passe à la suivante |
| `/stop`         | Arrête et vide la queue |
| `/queue`        | Affiche la file d'attente |
| `/pause`        | Pause la lecture |
| `/resume`       | Reprend la lecture |
| `/leave`        | Quitte le vocal |

## Roadmap

Voir [ROADMAP.md](ROADMAP.md)
