# ROADMAP — Inkyra Bot

> Bot Discord modulaire en Go — mono-guild, SQLite, Docker. Versioning : MINOR pour features, PATCH pour fixes/robustesse.

---

## Statut actuel — V1.0.1 ✅

| Système                | État                                                |
| ---------------------- | --------------------------------------------------- |
| Tickets                | ✅ Complet                                          |
| XP / Levels / Prestige | ✅ Complet — cooldowns persistés en SQLite          |
| Économie               | ✅ Complet                                          |
| Mini-jeux              | ✅ Complet — sessions blackjack purgées au restart  |
| Musique                | ✅ Docker/Linux — stub Windows                      |
| Achievements           | ✅ Complet — 14 achievements, déclencheurs branchés |
| Sécurité & robustesse  | ✅ Audit v1.0.1 appliqué                            |
| Tests                  | ❌ Aucun                                            |

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

## V1.0.2 — Améliorations UX & gameplay

- [ ] Anti-spam (rate limiting messages, mute temporaire automatique)
- [ ] Auto-roles (rôle attribué à l'arrivée, rôles par reaction/bouton)
- [ ] Logs de modération (kicks, bans, timeouts) dans un canal dédié
- [ ] Tests unitaires : `TotalXPForLevel`, `BJ.Payout`, daily streak, cooldown XP
- [ ] Achievements supplémentaires (streaks daily, prestige, wins en jeux)
- [ ] Notifications de level-up configurables (canal dédié ou DM)
- [ ] Cooldowns work/daily configurables via `guild_settings`
- [ ] Plafond de mise configurable dans les mini-jeux
- [ ] Embeds avancés et menus de sélection (select menus Discord)

---

## V1.0.3 — Robustesse & observabilité

- [ ] Structured logging (`slog`) avec niveaux DEBUG/INFO/ERROR
- [ ] Métriques basiques (commandes les plus utilisées, uptime)
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
