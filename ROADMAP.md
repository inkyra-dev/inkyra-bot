# ROADMAP — Inkyra Bot

> Bot Discord modulaire en Go — mono-guild, SQLite, Docker. Versioning : MINOR pour features, PATCH pour fixes/robustesse.

---

## Statut actuel — V1.0.3 ✅ (partiel)

| Système                | État |
| ---------------------- | ---- |
| Tickets                | ✅ Complet |
| XP / Levels / Prestige | ✅ Complet — cooldowns persistés en SQLite, canal configurable |
| Économie               | ✅ Complet — cooldowns daily/work configurables par serveur |
| Mini-jeux              | ✅ Complet — mise max configurable, achievements jeux |
| Musique                | ✅ Docker/Linux — stub Windows |
| Achievements           | ✅ Complet — 22 achievements, compteurs user_stats |
| Modération             | ✅ Anti-spam, auto-rôles, logs bans/timeouts |
| Configuration          | ✅ /config + 5 réglages par serveur |
| Logging structuré      | ✅ slog complet, JSON/text, niveaux configurables |
| Métriques              | ✅ Compteurs atomiques, /stats enrichi, top-5 commandes |
| Health check HTTP      | ✅ /health + /ready, HEALTH_PORT configurable |
| Validation config      | ✅ TOKEN/GUILD_ID fatals, LOG_LEVEL warn, DB_PATH write-check |
| Graceful shutdown      | ✅ Ordre garanti, timeout 10s, durée loggée |
| Tests                  | ❌ Aucun |

---

## V1.0 — Stabilisation & complétion ✅

### Fixes bloquants

- [x] Persist cooldowns XP — colonne `last_xp_at` en SQLite
- [x] Sessions blackjack — table `blackjack_sessions`, purge au démarrage
- [x] `guild_settings` — branchement config tickets sur la table DB

### Features incomplètes

- [x] `/econleaderboard` — slash command enregistrée
- [x] Achievements — repo + manager + `/achievements` + déclencheurs

### Qualité

- [x] Audit global post-fixes

---

## V1.0.1 — Corrections audit & robustesse ✅

> Objectif : corriger les failles et dettes identifiées à l'audit avant d'ajouter des features.

### 🔴 Corrections urgentes

- [x] Valider `amount > 0` dans les commandes admin money/xp — `admin_cmd.go`
- [x] Ajouter `AND guild_id=?` dans `items_repo.go` — `Buy` et `GetByID`
- [x] Corriger `PlayerHandStr(reveal bool)` — paramètre ignoré supprimé
- [x] Capturer les erreurs dans `cmdResetUser` — reset partiel visible pour l'admin

### 🟡 Dette technique

- [x] Dédupliquer les 4 blocs leaderboard → `respondLeaderboard` + `xpLeaderboardBody` + `econLeaderboardBody`
- [x] Nommer les magic numbers en constantes dans `economy/manager.go`
- [x] Supprimer `IsUnlocked()` dead code — `achievements_repo.go`
- [x] Signal d'arrêt sur `BJManager.cleanup()` et `music.Player.loop` (graceful shutdown)

---

## V1.0.2 — Améliorations UX & gameplay ✅

### 🛡️ Modération

- [x] Anti-spam — rate-limit >5 messages/5s, timeout Discord natif 5 min, warning, log dans `LOG_CHANNEL_ID`
- [x] Auto-rôles — rôle attribué automatiquement à l'arrivée (`/setautorole`)
- [x] Embed rôles interactif — `/setuproles` + `/addrolebutton` + toggle boutons en mémoire
- [x] Logs de modération — bans/débans loggués avec modérateur (audit log Discord), timeouts anti-spam loggués

### ⚙️ Configuration par serveur

- [x] Canal level-up configurable — `/setlevelupchannel`, annonces prestige redirigées
- [x] Cooldown `/daily` configurable — `/setdailycooldown` (1–168h, défaut 24h)
- [x] Cooldown `/work` configurable — `/setworkcooldown` (1–168h, défaut 4h)
- [x] Mise maximale jeux — `/setmaxbet` (0 = illimité), appliqué sur les 4 mini-jeux
- [x] `/config` — embed éphémère avec les 9 réglages du serveur

### 🏆 Achievements (14 → 22)

- [x] `msg_100`, `msg_500`, `msg_1000` — compteur messages dans `user_stats`
- [x] `bj_win_10`, `bj_win_50` — compteur victoires BJ dans `user_stats`
- [x] `slots_jackpot` — triple jackpot (×20) aux machines à sous
- [x] `dice_double6` — dé sur 6
- [x] `welcome` — premier message après l'arrivée sur le serveur

### 🎖️ Leaderboard unifié

- [x] `/leaderboard type:xp|economie` — une seule commande, pagination partagée
- [x] Helpers `showXPLeaderboard` / `showEconLeaderboard` — zéro duplication de rendu

---

## V1.0.3 — Robustesse & observabilité ✅ (partiel)

### 📋 Logging structuré

- [x] Migration complète `log.Printf` → `slog` (14 fichiers, zéro import `"log"` résiduel)
- [x] `internal/logger` — handler JSON/text configurable via `LOG_FORMAT`, niveau via `LOG_LEVEL`
- [x] Attribut `"component"` sur chaque log + contexte pertinent (`guild_id`, `user_id`, `command`, `error`)

### 📊 Métriques

- [x] `internal/metrics` — singleton Go pur, compteurs `atomic.Int64`, zéro dépendance externe
- [x] `IncrCommand(name)` / `IncrMessage()` / `IncrSpamTimeout()` / `IncrDBError()` instrumentés
- [x] `/stats` enrichi — uptime formaté, messages traités, top-5 commandes, erreurs DB

### 🏥 Health check HTTP

- [x] `GET /health` — JSON `{"status","uptime","guilds","version"}`
- [x] `GET /ready` — 200/503 selon état de la session Discord
- [x] `HEALTH_PORT` configurable (défaut 8080, 0 = désactivé)
- [x] Arrêt propre dans la séquence de shutdown

### 🔧 Validation config & shutdown

- [x] TOKEN/GUILD_ID → `slog.Error` + `os.Exit(1)` avec message explicite
- [x] LOG_LEVEL invalide → `slog.Warn` + fallback `info`
- [x] DB_PATH → write-check réel (`os.CreateTemp`)
- [x] Ordre shutdown garanti : health → `dg.Close` → `handler.Shutdown` → `db.Close` — timeout 10s
- [x] Constante `Version = "1.0.3"` dans `config/config.go`

### ❌ Reporté à V1.0.4

- [ ] Tests unitaires : `TotalXPForLevel`, `BJ.Payout`, daily streak, cooldown XP
- [ ] `IncrDBError` coverage complète (economy, tickets, BJ actuellement non couverts)

---

## V1.1 — Multi-guild & Dashboard _(long terme)_

> Pré-requis : V1.x stable, dashboard NextJS opérationnel.

- [ ] Support multi-guild (`guild_settings` devient source de vérité par guild)
- [ ] API REST interne (ou WebSocket) pour le dashboard admin
- [ ] Dashboard — stats (XP, économie, tickets actifs)
- [ ] Dashboard — configuration bot par guild (canaux, rôles, cooldowns)
- [ ] Dashboard — gestion items du shop
- [ ] Authentification dashboard via Discord OAuth2

---

## V1.2 — Vision SaaS _(très long terme)_

- [ ] Onboarding multi-serveurs autonome (invite link + setup guidé)
- [ ] Plans (free / premium) avec limites de features
- [ ] Site docs MDX public (`inkyra-docs`) complet et à jour
- [ ] Dashboard exposé aux membres selon leurs rôles Discord
