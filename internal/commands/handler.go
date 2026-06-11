package commands

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/achievements"
	"discord-bot/internal/config"
	"discord-bot/internal/database"
	"discord-bot/internal/economy"
	"discord-bot/internal/games"
	"discord-bot/internal/metrics"
	"discord-bot/internal/moderation"
	"discord-bot/internal/music"
	"discord-bot/internal/repositories"
	"discord-bot/internal/tickets"
	"discord-bot/internal/xp"
)

type CommandFunc func(s *discordgo.Session, i *discordgo.InteractionCreate)

type Handler struct {
	session     *discordgo.Session
	cfg         *config.Config
	db          *database.DB
	tickets     *tickets.Manager
	music       *music.Manager
	xpMgr       *xp.Manager
	econMgr     *economy.Manager
	gamesMgr    *games.Manager
	achMgr      *achievements.Manager
	statsRepo   *repositories.StatsRepo
	spamMonitor *moderation.Monitor
	rolesMgr    *moderation.RolesManager
	modLog      *moderation.Logger
	commands    map[string]CommandFunc
}

func NewHandler(s *discordgo.Session, cfg *config.Config, db *database.DB) *Handler {
	xpRepo := repositories.NewXPRepo(db)
	econRepo := repositories.NewEconomyRepo(db)
	itemRepo := repositories.NewItemsRepo(db)
	bjRepo := repositories.NewBJRepo(db)
	achRepo := repositories.NewAchievementRepo(db)
	statsRepo := repositories.NewStatsRepo(db)
	achMgr := achievements.NewManager(achRepo)

	modLog := moderation.NewLogger(cfg.LogChannelID)

	h := &Handler{
		session:     s,
		cfg:         cfg,
		db:          db,
		tickets:     tickets.NewManager(db, cfg),
		music:       music.NewManager(),
		xpMgr:       xp.NewManager(xpRepo, statsRepo, achMgr, db),
		econMgr:     economy.NewManager(econRepo, itemRepo, achMgr),
		gamesMgr:    games.NewManager(econRepo, bjRepo),
		achMgr:      achMgr,
		statsRepo:   statsRepo,
		spamMonitor: moderation.NewMonitor(modLog),
		rolesMgr:    moderation.NewRolesManager(db),
		modLog:      modLog,
		commands:    make(map[string]CommandFunc),
	}
	h.register()
	metrics.GetMetrics() // initialise le singleton (startTime) dès le démarrage du handler
	return h
}

func (h *Handler) register() {
	// Basique
	h.commands["ping"] = h.ping
	h.commands["help"] = h.help
	h.commands["stats"] = h.stats

	// Tickets
	h.commands["ticket"] = h.cmdTicket
	h.commands["close"] = h.cmdClose
	h.commands["adduser"] = h.cmdAddUser
	h.commands["removeuser"] = h.cmdRemoveUser

	// Musique
	h.commands["play"] = h.cmdPlay
	h.commands["skip"] = h.cmdSkip
	h.commands["stop"] = h.cmdStop
	h.commands["queue"] = h.cmdQueue
	h.commands["pause"] = h.cmdPause
	h.commands["resume"] = h.cmdResume
	h.commands["leave"] = h.cmdLeave

	// XP
	h.commands["rank"] = h.cmdRank
	h.commands["leaderboard"] = h.cmdLeaderboard
	h.commands["addxp"] = h.cmdAddXP
	h.commands["removexp"] = h.cmdRemoveXP
	h.commands["setlevel"] = h.cmdSetLevel
	h.commands["prestige"] = h.cmdPrestige

	// Économie
	h.commands["balance"] = h.cmdBalance
	h.commands["daily"] = h.cmdDaily
	h.commands["work"] = h.cmdWork
	h.commands["deposit"] = h.cmdDeposit
	h.commands["withdraw"] = h.cmdWithdraw
	h.commands["pay"] = h.cmdPay
	h.commands["shop"] = h.cmdShop
	h.commands["buy"] = h.cmdBuy
	h.commands["sell"] = h.cmdSell
	h.commands["inventory"] = h.cmdInventory
	h.commands["econleaderboard"] = h.cmdEconLeaderboard

	// Jeux
	h.commands["coinflip"] = h.cmdCoinflip
	h.commands["dice"] = h.cmdDice
	h.commands["slots"] = h.cmdSlots
	h.commands["blackjack"] = h.cmdBlackjack

	// Achievements
	h.commands["achievements"] = h.cmdAchievements

	// Configuration
	h.commands["setlevelupchannel"] = h.cmdSetLevelUpChannel
	h.commands["setdailycooldown"] = h.cmdSetDailyCooldown
	h.commands["setworkcooldown"] = h.cmdSetWorkCooldown
	h.commands["setmaxbet"] = h.cmdSetMaxBet
	h.commands["config"] = h.cmdConfig

	// Modération
	h.commands["setautorole"] = h.cmdSetAutoRole
	h.commands["setuproles"] = h.cmdSetupRoles
	h.commands["addrolebutton"] = h.cmdAddRoleButton

	// Admin
	h.commands["givemoney"] = h.cmdGiveMoney
	h.commands["removemoney"] = h.cmdRemoveMoney
	h.commands["resetuser"] = h.cmdResetUser
	h.commands["setxp"] = h.cmdSetXP
	h.commands["economy-reset"] = h.cmdEconomyReset
}

func (h *Handler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		name := i.ApplicationCommandData().Name
		metrics.GetMetrics().IncrCommand(name)
		if fn, ok := h.commands[name]; ok {
			fn(s, i)
		}
	case discordgo.InteractionMessageComponent:
		h.handleComponent(s, i)
	}
}

func (h *Handler) handleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	id := i.MessageComponentData().CustomID

	switch id {
	case "close_ticket":
		h.closeTicketComponent(s, i)
	case "transcript_ticket":
		h.transcriptTicketComponent(s, i)
	case "bj_hit":
		h.bjHit(s, i)
	case "bj_stand":
		h.bjStand(s, i)
	default:
		// Leaderboard pagination : "xplb:N" ou "ecolb:N"
		if strings.HasPrefix(id, "xplb:") {
			h.handleXPLeaderboardPage(s, i, id)
		} else if strings.HasPrefix(id, "ecolb:") {
			h.handleEcoLeaderboardPage(s, i, id)
		} else if strings.HasPrefix(id, "role_toggle:") {
			roleID := strings.TrimPrefix(id, "role_toggle:")
			h.rolesMgr.HandleRoleToggle(s, i, roleID)
		}
	}
}

func (h *Handler) handleXPLeaderboardPage(s *discordgo.Session, i *discordgo.InteractionCreate, id string) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return
	}

	var page int
	fmt.Sscanf(parts[1], "%d", &page)
	if page < 1 {
		page = 1
	}
	h.showXPLeaderboard(s, i, page, true)
}

func (h *Handler) RegisterCommands() {
	userOpt := func(name, desc string) *discordgo.ApplicationCommandOption {
		return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionUser, Name: name, Description: desc, Required: true}
	}
	intOpt := func(name, desc string, req bool) *discordgo.ApplicationCommandOption {
		return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionInteger, Name: name, Description: desc, Required: req}
	}
	strOpt := func(name, desc string, req bool) *discordgo.ApplicationCommandOption {
		return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionString, Name: name, Description: desc, Required: req}
	}

	defs := []*discordgo.ApplicationCommand{
		// ── Basique ──
		{Name: "ping", Description: "Répond pong 🏓"},
		{Name: "help", Description: "Liste toutes les commandes"},
		{Name: "stats", Description: "Statistiques du bot"},

		// ── Tickets ──
		{Name: "ticket", Description: "Ouvre un ticket de support"},
		{Name: "close", Description: "Ferme le ticket actuel"},
		{Name: "adduser", Description: "Ajoute un utilisateur au ticket",
			Options: []*discordgo.ApplicationCommandOption{userOpt("user", "Utilisateur à ajouter")}},
		{Name: "removeuser", Description: "Retire un utilisateur du ticket",
			Options: []*discordgo.ApplicationCommandOption{userOpt("user", "Utilisateur à retirer")}},

		// ── Musique ──
		{Name: "play", Description: "Joue une musique YouTube",
			Options: []*discordgo.ApplicationCommandOption{strOpt("query", "URL ou recherche YouTube", true)}},
		{Name: "skip", Description: "Passe à la musique suivante"},
		{Name: "stop", Description: "Arrête et vide la queue"},
		{Name: "queue", Description: "Affiche la file d'attente"},
		{Name: "pause", Description: "Met en pause"},
		{Name: "resume", Description: "Reprend la lecture"},
		{Name: "leave", Description: "Quitte le vocal"},

		// ── XP ──
		{Name: "rank", Description: "Affiche ton profil XP",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "Utilisateur (optionnel)", Required: false},
			}},
		{Name: "leaderboard", Description: "Classement XP ou économique du serveur",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "type",
					Description: "Type de classement (défaut : xp)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "XP", Value: "xp"},
						{Name: "Économie", Value: "economie"},
					},
				},
				intOpt("page", "Numéro de page", false),
			}},
		{Name: "prestige", Description: "Prestige (réinitialise XP au niveau 100)"},
		{Name: "addxp", Description: "[Admin] Ajoute de l'XP",
			Options: []*discordgo.ApplicationCommandOption{userOpt("user", "Cible"), intOpt("amount", "Quantité", true)}},
		{Name: "removexp", Description: "[Admin] Retire de l'XP",
			Options: []*discordgo.ApplicationCommandOption{userOpt("user", "Cible"), intOpt("amount", "Quantité", true)}},
		{Name: "setlevel", Description: "[Admin] Définit le niveau",
			Options: []*discordgo.ApplicationCommandOption{userOpt("user", "Cible"), intOpt("level", "Niveau", true)}},

		// ── Économie ──
		{Name: "balance", Description: "Affiche ton solde",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "Utilisateur (optionnel)", Required: false},
			}},
		{Name: "daily", Description: "Récompense quotidienne"},
		{Name: "work", Description: "Travailler pour gagner des coins (cd: 4h)"},
		{Name: "deposit", Description: "Déposer des coins en banque",
			Options: []*discordgo.ApplicationCommandOption{strOpt("amount", "Montant ou 'all'", true)}},
		{Name: "withdraw", Description: "Retirer des coins de la banque",
			Options: []*discordgo.ApplicationCommandOption{strOpt("amount", "Montant ou 'all'", true)}},
		{Name: "pay", Description: "Payer un autre utilisateur",
			Options: []*discordgo.ApplicationCommandOption{userOpt("user", "Destinataire"), intOpt("amount", "Montant", true)}},
		{Name: "shop", Description: "Affiche la boutique"},
		{Name: "buy", Description: "Acheter un item",
			Options: []*discordgo.ApplicationCommandOption{intOpt("id", "ID de l'item", true)}},
		{Name: "sell", Description: "Revendre un item",
			Options: []*discordgo.ApplicationCommandOption{intOpt("id", "ID de l'item", true)}},
		{Name: "inventory", Description: "Ton inventaire"},
		{Name: "econleaderboard", Description: "Classement économique du serveur",
			Options: []*discordgo.ApplicationCommandOption{intOpt("page", "Numéro de page", false)}},

		// ── Jeux ──
		{Name: "coinflip", Description: "Pile ou face",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionString, Name: "choix", Description: "pile ou face", Required: true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Pile", Value: "pile"},
						{Name: "Face", Value: "face"},
					}},
				intOpt("mise", "Mise en coins", true),
			}},
		{Name: "dice", Description: "Lance un dé",
			Options: []*discordgo.ApplicationCommandOption{intOpt("mise", "Mise en coins", true)}},
		{Name: "slots", Description: "Machines à sous",
			Options: []*discordgo.ApplicationCommandOption{intOpt("mise", "Mise en coins", true)}},
		{Name: "blackjack", Description: "Jouer au blackjack",
			Options: []*discordgo.ApplicationCommandOption{intOpt("mise", "Mise en coins", true)}},
		{Name: "achievements", Description: "Affiche les achievements d'un utilisateur",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "Utilisateur (optionnel)", Required: false},
			}},

		// ── Configuration ──
		{Name: "setlevelupchannel", Description: "[Admin] Définit le canal des annonces de level-up",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionChannel, Name: "canal", Description: "Canal d'annonce", Required: true},
			}},
		{Name: "setdailycooldown", Description: "[Admin] Définit le cooldown de /daily (1–168h)",
			Options: []*discordgo.ApplicationCommandOption{intOpt("heures", "Nombre d'heures (défaut: 24)", true)}},
		{Name: "setworkcooldown", Description: "[Admin] Définit le cooldown de /work (1–168h)",
			Options: []*discordgo.ApplicationCommandOption{intOpt("heures", "Nombre d'heures (défaut: 4)", true)}},
		{Name: "setmaxbet", Description: "[Admin] Définit la mise maximale pour les jeux (0 = illimité)",
			Options: []*discordgo.ApplicationCommandOption{intOpt("montant", "Montant (0 = illimité)", true)}},
		{Name: "config", Description: "[Admin] Affiche la configuration actuelle du serveur"},

		// ── Modération ──
		{Name: "setautorole", Description: "[Admin] Définit le rôle automatique à l'arrivée d'un membre",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionRole, Name: "role", Description: "Rôle à attribuer", Required: true},
			}},
		{Name: "setuproles", Description: "[Admin] Poste l'embed de sélection de rôles dans ce canal"},
		{Name: "addrolebutton", Description: "[Admin] Ajoute un bouton à l'embed de rôles",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionRole, Name: "role", Description: "Rôle à toggler", Required: true},
				strOpt("label", "Texte du bouton", true),
				strOpt("emoji", "Emoji Unicode (optionnel)", false),
			}},

		// ── Admin ──
		{Name: "givemoney", Description: "[Admin] Donne des coins",
			Options: []*discordgo.ApplicationCommandOption{userOpt("user", "Cible"), intOpt("amount", "Montant", true)}},
		{Name: "removemoney", Description: "[Admin] Retire des coins",
			Options: []*discordgo.ApplicationCommandOption{userOpt("user", "Cible"), intOpt("amount", "Montant", true)}},
		{Name: "resetuser", Description: "[Admin] Réinitialise XP + économie",
			Options: []*discordgo.ApplicationCommandOption{userOpt("user", "Cible")}},
		{Name: "setxp", Description: "[Admin] Définit l'XP total",
			Options: []*discordgo.ApplicationCommandOption{userOpt("user", "Cible"), intOpt("amount", "XP total", true)}},
		{Name: "economy-reset", Description: "[Admin] Réinitialise l'économie",
			Options: []*discordgo.ApplicationCommandOption{userOpt("user", "Cible")}},
	}

	for _, cmd := range defs {
		if _, err := h.session.ApplicationCommandCreate(h.session.State.User.ID, h.cfg.GuildID, cmd); err != nil {
			slog.Error("création commande échouée", "component", "commands", "command", cmd.Name, "error", err)
		}
	}
	slog.Info("commandes enregistrées", "component", "commands", "count", len(defs))
}

func (h *Handler) Shutdown() {
	h.gamesMgr.Shutdown()
	h.music.Shutdown()
	h.spamMonitor.Shutdown()
	slog.Info("bot arrêté proprement", "component", "commands")
}

// HandleVoiceStateUpdate est appelé par l'event voice.
func (h *Handler) HandleVoiceStateUpdate(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	h.music.HandleVoiceStateUpdate(s, vs)
}

// HandleMessageCreate est appelé par l'event message pour le gain d'XP et l'anti-spam.
func (h *Handler) HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || m.GuildID == "" {
		return
	}
	h.spamMonitor.Check(s, m.GuildID, m.ChannelID, m.Author.ID, m.Author.Username)
	h.xpMgr.HandleMessage(s, m.GuildID, m.Author.ID, m.ChannelID)
}

// HandleMemberJoin est appelé quand un membre rejoint le serveur.
func (h *Handler) HandleMemberJoin(s *discordgo.Session, guildID, userID string) {
	h.rolesMgr.HandleMemberJoin(s, guildID, userID)
	h.achMgr.Check(guildID, userID, "welcome")
}

// HandleBanAdd est appelé quand un membre est banni.
func (h *Handler) HandleBanAdd(s *discordgo.Session, guildID, userID string) {
	h.modLog.LogBan(s, guildID, userID)
}

// HandleBanRemove est appelé quand un ban est levé.
func (h *Handler) HandleBanRemove(s *discordgo.Session, guildID, userID string) {
	h.modLog.LogUnban(s, guildID, userID)
}
