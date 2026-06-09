# ROADMAP — Inkyra Bot

> Bot Discord modulaire en Go — mono-guild, SQLite, Docker. Versioning : `1.0.1` MINOR pour features, PATCH pour fixes.

---

## Statut actuel — V1.0 (70%)

| Système                | État                                                |
| ---------------------- | --------------------------------------------------- |
| Tickets                | ✅ Complet                                          |
| XP / Levels / Prestige | ✅ Complet — cooldowns persistés en SQLite          |
| Économie               | ✅ Complet                                          |
| Mini-jeux              | ✅ Complet — sessions blackjack purgées au restart  |
| Musique                | ✅ Docker/Linux — stub Windows                      |
| Achievements           | ✅ Complet — 14 achievements, déclencheurs branchés |
| /econleaderboard       | ✅ Complet                                          |
| guild_settings         | ⚠️ Partiellement branché — migration incomplète     |
| Tests                  | ❌ Aucun                                            |

---

## V1.0 — Stabilisation & complétion _(en cours)_

> Objectif : un bot stable, sans données perdues au restart, avec toutes les features annoncées fonctionnelles.

### Fixes bloquants

- [x] **Persist cooldowns XP** — colonne `last_xp_at` en SQLite
- [x] **Sessions blackjack** — table `blackjack_sessions`, purge au démarrage
- [x] **guild_settings** — branchement config tickets sur la table DB

### Features incomplètes

- [x] **`/econleaderboard`** — slash command enregistrée
- [x] **Achievements** — repo + manager + `/achievements` + déclencheurs

### Qualité

- [ ] Tests unitaires : `TotalXPForLevel`, `BJ.Payout`, daily streak, cooldown XP
- [x] Audit global post-fixes

---

## V1.0.1 — Corrections audit & robustesse

> Objectif : corriger les failles et dettes identifiées à l'audit avant d'ajouter des features.

### 🔴 Corrections urgentes (audit)

- [ ] Valider `amount > 0` dans les commandes admin money/xp — `admin_cmd.go:38,53,94`
- [ ] Ajouter `AND guild_id=?` dans `items_repo.go` — `Buy` et `GetByID`
- [ ] Corriger `PlayerHandStr(reveal bool)` — paramètre ignoré, `blackjack.go:127`
- [ ] Capturer les erreurs dans `cmdResetUser` — `admin_cmd.go:78`

### 🟡 Dette technique (audit)

- [ ] Dédupliquer les 4 blocs leaderboard → helper `renderLeaderboardPage`
- [ ] Nommer les magic numbers en constantes dans `economy/manager.go`
- [ ] Supprimer `IsUnlocked()` dead code — `achievements_repo.go:47`
- [ ] Signal d'arrêt sur `BJManager.cleanup()` et `music.Player.loop` (graceful shutdown)

### Modération & automatismes

- [ ] Anti-spam (rate limiting messages, mute temporaire automatique)
- [ ] Auto-roles (rôle attribué à l'arrivée, rôles par reaction/bouton)
- [ ] Embeds avancés et menus de sélection (select menus Discord)
- [ ] Logs de modération (kicks, bans, timeouts) dans un canal dédié

---

## V1.0.2 — Améliorations UX & gameplay

- [ ] Achievements supplémentaires (streaks daily, prestige, wins en jeux)
- [ ] Notifications de level-up configurables (canal dédié ou DM)
- [ ] Cooldowns work/daily configurables via guild_settings
- [ ] Pagination améliorée (leaderboards XP + économie unifiés)
- [ ] Plafond de mise configurable dans les mini-jeux

---

## V1.0.3 — Robustesse & observabilité

- [ ] Structured logging (`slog`) avec niveaux DEBUG/INFO/ERROR
- [ ] Métriques basiques (commandes les plus utilisées, uptime)
- [ ] Graceful shutdown complet (drain des sessions actives avant arrêt)
- [ ] Health check endpoint HTTP minimal (pour Uptime Kuma)
- [ ] Validation config au démarrage — message clair si TOKEN vide

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
