```text
discord-bot/
├── cmd/main.go                     ← entrée, wiring de tout
├── internal/
├── repositories/
│   ├── xp_repo.go        ← CRUD XP, leaderboard, rang
│   ├── economy_repo.go   ← wallet/bank, dépôt/retrait, transfert atomique (transactions SQL)
│   └── items_repo.go     ← shop, achat/vente atomique, inventaire
├── xp/
│   ├── calculator.go     ← formule MEE6 (5n²+50n+100), progress bar unicode
│   └── manager.go        ← cooldown mémoire, gain auto XP, level-up, prestige
├── economy/
│   └── manager.go        ← daily (streak bonus), work (4h cd, job aléatoire), deposit/withdraw all
├── games/
│   ├── games.go          ← coinflip, dice (×0-3), slots (7 symboles, combinaisons)
│   ├── blackjack.go      ← moteur BJ complet (hit/stand/bust/push/blackjack naturel)
│   └── manager.go        ← BJManager sessions mémoire + cleanup goroutine
└── commands/
    ├── helpers.go         ← formatCoins, parseCoins, leaderboardButtons
    ├── xp_cmd.go          ← /rank /leaderboard /addxp /removexp /setlevel /prestige
    ├── economy_cmd.go     ← /balance /daily /work /deposit /withdraw /pay /shop /buy /sell /inventory
    ├── games_cmd.go       ← /coinflip /dice /slots /blackjack (boutons Hit/Stand interactifs)
    └── admin_cmd.go       ← /givemoney /removemoney /resetuser /setxp /economy-reset

├── config/config.go            ← chargement .env
│   ├── database/
│   │   ├── database.go             ← connexion SQLite (modernc, pure Go)
│   │   ├── migrations.go           ← CREATE TABLE automatiques
│   │   └── tickets_repo.go         ← CRUD tickets
│   ├── utils/
│   │   ├── embeds.go               ← helpers embed + couleurs
│   │   └── respond.go              ← Respond / RespondEmbed / RespondEphemeral
│   ├── events/
│   │   ├── ready.go                ← enregistrement commandes au démarrage
│   │   ├── interaction.go          ← routing slash commands + boutons
│   │   └── voice.go                ← cleanup musique si vocal vide
│   ├── commands/
│   │   ├── handler.go              ← Handler central, RegisterCommands()
│   │   ├── basic.go                ← /ping /help /stats
│   │   ├── ticket_cmd.go           ← /ticket /close /adduser /removeuser + boutons
│   │   └── music_cmd.go            ← /play /skip /stop /queue /pause /resume /leave
│   ├── tickets/
│   │   ├── manager.go              ← Open / Close / AddUser / RemoveUser
│   │   └── transcript.go           ← génération transcript TXT
│   └── music/
│       ├── queue.go                ← Queue thread-safe
│       ├── player.go               ← Player par guild (loop, skip, stop)
│       ├── manager.go              ← Manager multi-guild + cleanup vocal
│       ├── ytdlp.go                ← GetInfo() via yt-dlp
│       ├── audio_cgo.go            ← playAudioFile (avec dgvoice, build:cgo)
│       └── audio_nocgo.go          ← stub silencieux (build:!cgo, dev Windows)
├── Dockerfile                      ← builder Go + image debian avec ffmpeg/yt-dlp
├── docker-compose.yml
└── .env                            ← TOKEN + GUILD_ID + optionnels
```

# Dev Windows (musique stub, tout le reste fonctionnel)

go run ./cmd/

# Docker Linux (musique complète avec ffmpeg + libopus)

docker compose up -d --build

# Variables .env optionnelles à remplir pour les tickets :

STAFF_ROLE_ID= # rôle staff qui voit les tickets
TICKET_CATEGORY_ID= # catégorie où créer les salons ticket-xxx
LOG_CHANNEL_ID= # salon où envoyer les transcripts à la fermeture

Points clés :

- Transactions SQL sur tous les transferts financiers (pas de double-dépense)
- Cache mémoire pour les cooldowns XP (pas de hit DB à chaque message)
- Blackjack multi-étapes avec boutons Discord (sessions expirées après 10 min)
- Leaderboard paginé avec boutons Précédent/Suivant
- Permissions admin vérifiées via l'API Discord (rôle Administrator)
- /deposit all / /withdraw all supportés
- Variables .env inchangées — rien de nouveau requis, tout fonctionne avec TOKEN + GUILD_ID.
