# Architecture — Inkyra Bot

## Vue d'ensemble

```
cmd/main.go
    │
    ├── internal/events/          ← adaptateurs fins discordgo
    │       ready.go              → RegisterCommands() au démarrage
    │       interaction.go        → délègue à Handler.Handle()
    │       voice.go              → délègue à Handler.HandleVoiceStateUpdate()
    │       member.go             → délègue à Handler.HandleMemberJoin()
    │       modlog.go             → délègue à Handler.HandleBanAdd/Remove()
    │
    └── internal/commands/Handler ← hub central
            │
            ├── internal/xp/Manager
            │       ├── internal/repositories/XPRepo
            │       ├── internal/repositories/StatsRepo
            │       └── internal/database/DB       (ChannelSettings interface)
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
            ├── internal/moderation/Monitor        (anti-spam, in-memory)
            ├── internal/moderation/Logger         (logs embed → Discord)
            ├── internal/moderation/RolesManager   (auto-rôles, embed boutons)
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
| `internal/repositories` | SQL CRUD : `xp_repo`, `economy_repo`, `items_repo`, `bj_repo`, `achievements_repo`, `stats_repo` |
| `internal/xp` | Formule MEE6 (`5n²+50n+100`), cooldown XP persisté en DB (`last_xp_at`), prestige niveau 100, canal level-up configurable |
| `internal/economy` | Daily (streak), work, dépôt/retrait, transferts, shop — cooldowns injectés en paramètre depuis `guild_settings` |
| `internal/games` | Coinflip, dice, slots ; blackjack complet (Hit/Stand/Bust/Push/Natural) ; `enforceMaxBet` vérifie le plafond par guild |
| `internal/achievements` | Catalogue statique `All []Achievement` (22 entries) ; `Manager.Check()` fire-and-forget (échecs logués, jamais propagés) |
| `internal/moderation` | Anti-spam (rate-limit + timeout Discord), auto-rôles (GuildMemberAdd + boutons toggle), logs bans/timeouts dans `LOG_CHANNEL_ID` |
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
- `handleComponent()` — boutons ticket (`close_ticket`, `transcript_ticket`), blackjack (`bj_hit`, `bj_stand`), pagination leaderboard (`xplb:N`, `ecolb:N`), toggle rôles (`role_toggle:<roleID>`)

**Ajouter une slash command :** 1) `register()` dans `handler.go`, 2) `ApplicationCommand` dans `RegisterCommands()`, 3) méthode dans le `*_cmd.go` approprié.

### Helpers leaderboard (`xp_cmd.go` / `economy_cmd.go`)

La pagination XP et économie passe par des helpers privés partagés entre commandes et boutons :

- `respondLeaderboard(...)` — construit l'embed + boutons et appelle `InteractionRespond` (nouveau message ou update selon `update bool`)
- `showXPLeaderboard(s, i, page, update)` — lit le repo XP + appelle `respondLeaderboard`
- `showEconLeaderboard(s, i, page, update)` — lit le repo éco + appelle `respondLeaderboard`
- `xpLeaderboardBody(...)` — formate `[]LeaderboardEntry` → string markdown
- `econLeaderboardBody(...)` — formate `[]UserEconomy` → string markdown

`/leaderboard type:xp|economie` et `/econleaderboard` dispatche vers les mêmes helpers ; les boutons de pagination font de même — zéro duplication.

### Événements discordgo branchés

| Événement | Handler | Rôle |
|---|---|---|
| `MessageCreate` | `events.MessageCreate` | XP + anti-spam (antispam.Monitor.Check) |
| `InteractionCreate` | `events.InteractionCreate` | Slash commands + boutons |
| `VoiceStateUpdate` | `events.VoiceStateUpdate` | Cleanup musique |
| `GuildMemberAdd` | `events.GuildMemberAdd` | Auto-rôle + achievement `welcome` |
| `GuildBanAdd` | `events.GuildBanAdd` | Log ban dans `LOG_CHANNEL_ID` |
| `GuildBanRemove` | `events.GuildBanRemove` | Log déban dans `LOG_CHANNEL_ID` |

---

## Modération

### Anti-spam (`internal/moderation/antispam.go`)

- `Monitor` maintient un compteur `map[chanKey]int` (guild+channel+user) en mémoire avec `sync.Mutex`
- Un `time.Ticker` à 5 secondes remet tous les compteurs à zéro
- Au 6e message dans la fenêtre : timeout Discord natif de 5 min (`GuildMemberTimeout`), message éphémère d'avertissement (une seule fois par fenêtre via `warned map[userKey]bool`), log dans `LOG_CHANNEL_ID`
- `Shutdown()` ferme le canal `done` pour arrêter le ticker proprement

### Logs de modération (`internal/moderation/modlog.go`)

- `Logger{logChanID string}` — `LogTimeout`, `LogBan`, `LogUnban`
- `LogBan`/`LogUnban` : goroutine avec 600ms de délai pour laisser Discord peupler l'audit log, puis `GuildAuditLog` pour récupérer le modérateur
- Cast explicite : `s.GuildAuditLog(guildID, "", "", int(actionType), 5)` — `discordgo.AuditLogAction` est un type nommé, pas `int`

### Auto-rôles (`internal/moderation/autoroles.go`)

- `RolesManager` lit `auto_role_id` depuis `guild_settings` et appelle `GuildMemberRoleAdd` à chaque `GuildMemberAdd`
- `/setuproles` poste un embed épinglé dans le canal courant et sauvegarde `roles_channel_id` + `roles_message_id` en DB
- `/addrolebutton` insère dans `role_buttons` (UNIQUE guild+role) et met à jour l'embed via `refreshEmbed`
- Boutons : `role_toggle:<roleID>` — vérifie `i.Member.Roles`, ajoute ou retire le rôle, répond éphémère
- `ComponentEmoji` est un pointeur dans discordgo v0.29.0 : `&discordgo.ComponentEmoji{Name: emoji}`

---

## Base de données

SQLite via `modernc.org/sqlite` (pure Go, zéro CGO). Migrations auto au démarrage dans `database/migrations.go`.

| Table | Contenu |
|---|---|
| `tickets` | Tickets ouverts/fermés par canal |
| `guild_settings` | Config par serveur — staff role, catégorie tickets, log channel, auto_role_id, roles_channel_id, roles_message_id, levelup_channel_id, daily_cooldown_hours, work_cooldown_hours, max_bet |
| `user_xp` | XP total, niveau, prestige, `last_xp_at` (cooldown persisté) |
| `economy` | Wallet, bank, streak daily, `last_daily`, `last_work` |
| `items` | Catalogue boutique par guild |
| `inventory` | Possession items par user/guild |
| `achievements` | Succès débloqués (UNIQUE user/guild/key — idempotent) |
| `blackjack_sessions` | Sessions actives (purgées à chaque restart) |
| `role_buttons` | Boutons toggle de l'embed rôles (UNIQUE guild+role) |
| `user_stats` | Compteurs achievements : `message_count`, `bj_wins` (PRIMARY KEY guild+user) |

**Transactions SQL** sur tous les transferts financiers (deposit, withdraw, pay, buy, sell) pour prévenir les double-dépenses.

**Migrations idempotentes** : chaque ajout de colonne est précédé d'une vérification `pragma_table_info` — le démarrage est sûr peu importe la version de la DB existante.

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
    ├── moderation.Monitor.Shutdown() → close(done)
    │       └── resetLoop() goroutine : <-done → return
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
2. Si l'achievement nécessite un compteur (ex: `msg_100`), l'incrémenter via `StatsRepo.IncrementMessages` ou `IncrementBJWins` et déclencher `Check()` sur le seuil atteint
3. Appeler `achMgr.Check(guildID, userID, "ma_clé")` au bon endroit
4. Aucune migration nécessaire pour la table `achievements` — clé texte générique
5. Si un nouveau compteur est requis, ajouter la colonne dans `user_stats` via une migration idempotente dans `migrations.go`
