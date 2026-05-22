package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/config"
	"discord-bot/internal/database"
	"discord-bot/internal/economy"
	"discord-bot/internal/games"
	"discord-bot/internal/music"
	"discord-bot/internal/repositories"
	"discord-bot/internal/tickets"
	"discord-bot/internal/utils"
	"discord-bot/internal/xp"
)

type CommandFunc func(s *discordgo.Session, i *discordgo.InteractionCreate)

type Handler struct {
	session  *discordgo.Session
	cfg      *config.Config
	db       *database.DB
	tickets  *tickets.Manager
	music    *music.Manager
	xpMgr   *xp.Manager
	econMgr  *economy.Manager
	gamesMgr *games.Manager
	commands map[string]CommandFunc
}

func NewHandler(s *discordgo.Session, cfg *config.Config, db *database.DB) *Handler {
	xpRepo := repositories.NewXPRepo(db)
	econRepo := repositories.NewEconomyRepo(db)
	itemRepo := repositories.NewItemsRepo(db)

	h := &Handler{
		session:  s,
		cfg:      cfg,
		db:       db,
		tickets:  tickets.NewManager(db, cfg),
		music:    music.NewManager(),
		xpMgr:   xp.NewManager(xpRepo),
		econMgr:  economy.NewManager(econRepo, itemRepo),
		gamesMgr: games.NewManager(econRepo),
		commands: make(map[string]CommandFunc),
	}
	h.register()
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

	// Jeux
	h.commands["coinflip"] = h.cmdCoinflip
	h.commands["dice"] = h.cmdDice
	h.commands["slots"] = h.cmdSlots
	h.commands["blackjack"] = h.cmdBlackjack

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

	const pageSize = 10
	entries, total, err := h.xpMgr.GetLeaderboard(i.GuildID, page)
	if err != nil || len(entries) == 0 {
		utils.RespondEphemeral(s, i.Interaction, "Aucune donnée.")
		return
	}
	totalPages := (int(total) + pageSize - 1) / pageSize

	medals := []string{"🥇", "🥈", "🥉"}
	var sb strings.Builder
	for _, e := range entries {
		medal := ""
		if int(e.Rank)-1 < len(medals) {
			medal = medals[e.Rank-1] + " "
		}
		name := usernameOrID(s, i.GuildID, e.UserID)
		sb.WriteString(fmt.Sprintf("%s`#%d` **%s** — Niv. %d | %d XP\n", medal, e.Rank, name, e.Level, e.TotalXP))
	}

	embed := utils.EmbedFields("🏆 Classement XP", sb.String(), utils.ColorPurple,
		utils.Field("Page", fmt.Sprintf("%d / %d", page, totalPages), true))
	buttons := leaderboardButtons(page, totalPages, "xplb")

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: buttons,
		},
	})
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
		{Name: "leaderboard", Description: "Classement XP du serveur",
			Options: []*discordgo.ApplicationCommandOption{intOpt("page", "Numéro de page", false)}},
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
			log.Printf("[commands] Erreur création /%s: %v", cmd.Name, err)
		}
	}
	log.Printf("[commands] %d commandes enregistrées", len(defs))
}

// HandleVoiceStateUpdate est appelé par l'event voice.
func (h *Handler) HandleVoiceStateUpdate(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	h.music.HandleVoiceStateUpdate(s, vs)
}

// HandleMessageCreate est appelé par l'event message pour le gain d'XP.
func (h *Handler) HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || m.GuildID == "" {
		return
	}
	h.xpMgr.HandleMessage(s, m.GuildID, m.Author.ID, m.ChannelID)
}
