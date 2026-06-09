# Architecture — Inkyra Bot

## Vue d'ensemble

```
cmd/main.go
    │
    ├── internal/events/          ← adaptateurs fins discordgo
    │       ready.go              → RegisterCommands() au démarrage
    │       interaction.go        → délègue à Handler.Handle()
    │       voice.go              → délègue à Handler.HandleVoiceStateUpdate()
    │
    └── internal/commands/Handler ← hub central
            │
            ├── internal/xp/Manager
            │       └── internal/repositories/XPRepo
            │
            ├── internal/economy/Manager
            │       ├── internal/repositories/EconomyRepo
            │       └── internal/repositories/ItemsRepo
            │
            ├── internal/games/Manager
            │       ├── internal/games/BJManager   (sessions mémoire)
            │       └── internal/repositories/BJRepo
            │
            ├── internal/achievements/Manager
            │       └── internal/repositories/AchievementRepo
            │
            ├── internal/tickets/Manager
            │       └── internal/database/DB       (CRUD tickets)
            │
            └── internal/music/Manager
                    └── internal/music/Player × guild
```

Toutes les couches communiquent **vers le bas uniquement** : `commands` → `managers` → `repositories` → `database`. La couche `repositories` ne connaît pas `commands` ni les managers.

---

## Packages

| Package | Responsabilité |
|---|---|
| `cmd` | Wiring : config → DB → Handler → discordgo → signal SIGINT |
| `internal/config` | Chargement `.env` via godotenv |
| `internal/database` | Connexion SQLite (modernc pure Go), WAL mode, auto-migrations au démarrage |
| `internal/repositories` | SQL CRUD : `xp_repo`, `economy_repo`, `items_repo`, `bj_repo`, `achievements_repo` |
| `internal/xp` | Formule MEE6 (`5n²+50n+100`), cooldown XP persisté en DB (`last_xp_at`), prestige au niveau 100 |
| `internal/economy` | Daily (streak), work (4h cd), dépôt/retrait, transferts, shop — paramètres en constantes nommées |
| `internal/games` | Coinflip, dice, slots ; blackjack complet (Hit/Stand/Bust/Push/Natural) avec sessions mémoire |
| `internal/achievements` | Catalogue statique `All []Achievement` ; `Manager.Check()` fire-and-forget (échecs logués, jamais propagés) |
| `internal/tickets` | Ouverture/fermeture canaux Discord, permissions par overwrite, transcripts TXT |
| `internal/music` | `Player` par guild (boucle de lecture + goroutine audio) ; `Manager` multi-guild + cleanup vocal vide |
| `internal/events` | Adaptateurs fins entre discordgo et `commands.Handler` (interfaces, pas de logique) |
| `internal/commands` | Routing central (`Handler`), toutes les slash commands, helpers partagés |
| `internal/utils` | Builders d'embeds, constantes de couleurs, helpers `Respond*` |

---

## Routing des commandes

`internal/commands/handler.go` est le hub unique :

- `register()` — table `map[string]CommandFunc` peuplée au démarrage
- `Handle()` — dispatch par nom pour les slash commands, par `CustomID` pour les boutons
- `handleComponent()` — boutons ticket (`close_ticket`, `transcript_ticket`), blackjack (`bj_hit`, `bj_stand`), pagination leaderboard (`xplb:N`, `ecolb:N`)

**Ajouter une slash command :** 1) `register()` dans `handler.go`, 2) `ApplicationCommand` dans `RegisterCommands()`, 3) méthode dans le `*_cmd.go` approprié.

### Helpers leaderboard (`helpers.go`)

La pagination XP et économie passe par trois fonctions partagées :

- `respondLeaderboard(...)` — construit l'embed + boutons et appelle `InteractionRespond` (nouveau message ou update selon `update bool`)
- `xpLeaderboardBody(...)` — formate `[]LeaderboardEntry` → string markdown
- `econLeaderboardBody(...)` — formate `[]UserEconomy` → string markdown

---

## Base de données

SQLite via `modernc.org/sqlite` (pure Go, zéro CGO). Migrations auto au démarrage dans `database/migrations.go`.

| Table | Contenu |
|---|---|
| `tickets` | Tickets ouverts/fermés par canal |
| `guild_settings` | Config par serveur (staff role, catégorie, log channel) |
| `user_xp` | XP total, niveau, prestige, `last_xp_at` (cooldown persisté) |
| `economy` | Wallet, bank, streak daily, `last_daily`, `last_work` |
| `items` | Catalogue boutique par guild |
| `inventory` | Possession items par user/guild |
| `achievements` | Succès débloqués (UNIQUE user/guild/key — idempotent) |
| `blackjack_sessions` | Sessions actives (purgées à chaque restart) |

**Transactions SQL** sur tous les transferts financiers (deposit, withdraw, pay, buy, sell) pour prévenir les double-dépenses.

---

## Invariants clés

### Sécurité
- Toutes les requêtes SQL utilisent des paramètres positionnels `?` — zéro injection SQL possible
- `items_repo.GetByID` et `Buy` filtrent par `guild_id` — pas d'achat cross-guild
- Les commandes admin valident `amount > 0` (ou `>= 0` pour setxp) avant d'appeler la couche métier
- La vérification admin passe par le bit `PermissionAdministrator` Discord, pas un rôle hardcodé

### Achievements
`Manager.Check(guildID, userID, key)` est fire-and-forget : `INSERT OR IGNORE` en DB, erreurs logées et jamais propagées — les achievements ne peuvent pas bloquer le gameplay.

### Build tags CGO/musique
`audio_cgo.go` (tag `cgo`) appelle `dgvoice.PlayAudioFile`, requiert gcc + libopus.
`audio_nocgo.go` (tag `!cgo`) est un stub silencieux — utilisé sur Windows sans gcc.
La musique est fonctionnelle uniquement en Docker (image Linux avec ffmpeg + libopus).

### Graceful shutdown

À réception de SIGINT/SIGTERM, `cmd/main.go` appelle `handler.Shutdown()` avant que les defers (`dg.Close()`, `db.Close()`) ne s'exécutent :

```
handler.Shutdown()
    ├── games.Manager.Shutdown()
    │       └── BJManager.Shutdown() → close(done)
    │               └── cleanup() goroutine : <-done → return
    └── music.Manager.Shutdown()
            └── Player.Stop() × N guilds actifs
                    └── loop() goroutine : <-p.stop → return
```

---

## Ajouter un achievement

1. Ajouter une entrée dans `All []Achievement` dans `internal/achievements/manager.go`
2. Appeler `achMgr.Check(guildID, userID, "ma_clé")` au bon endroit dans le manager concerné
3. Aucune migration nécessaire — la table `achievements` est générique (clé texte)
