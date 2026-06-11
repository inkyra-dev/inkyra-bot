package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/metrics"
	"discord-bot/internal/utils"
)

func (h *Handler) ping(s *discordgo.Session, i *discordgo.InteractionCreate) {
	utils.Respond(s, i.Interaction, "pong 🏓")
}

func (h *Handler) help(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := utils.EmbedFields(
		"Commandes disponibles", "Voici toutes les commandes du bot.", utils.ColorBlue,
		utils.Field("/ping", "Répond pong 🏓", true),
		utils.Field("/help", "Affiche cette aide", true),
		utils.Field("/stats", "Stats du bot", true),
		utils.Field("/ticket", "Ouvre un ticket de support", true),
		utils.Field("/close", "Ferme le ticket actuel", true),
		utils.Field("/adduser `<user>`", "Ajoute un user au ticket", true),
		utils.Field("/removeuser `<user>`", "Retire un user du ticket", true),
		utils.Field("/play `<query>`", "Joue une musique YouTube", true),
		utils.Field("/skip", "Passe à la suivante", true),
		utils.Field("/stop", "Arrête et vide la queue", true),
		utils.Field("/queue", "Affiche la file d'attente", true),
		utils.Field("/pause", "Met en pause", true),
		utils.Field("/resume", "Reprend la lecture", true),
		utils.Field("/leave", "Quitte le vocal", true),
	)
	utils.RespondEmbed(s, i.Interaction, embed)
}

func (h *Handler) stats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	snap := metrics.GetMetrics().Snapshot()

	topCmds := "—"
	if len(snap.TopCommands) > 0 {
		parts := make([]string, len(snap.TopCommands))
		for j, c := range snap.TopCommands {
			parts[j] = fmt.Sprintf("`/%s` × %d", c.Name, c.Count)
		}
		topCmds = strings.Join(parts, "\n")
	}

	embed := utils.EmbedFields(
		"📊 Statistiques du bot", "", utils.ColorGreen,
		utils.Field("Uptime", metrics.FormatUptime(snap.Uptime), true),
		utils.Field("Serveurs", fmt.Sprintf("%d", len(s.State.Guilds)), true),
		utils.Field("Latence", fmt.Sprintf("%dms", s.HeartbeatLatency().Milliseconds()), true),
		utils.Field("Messages traités", fmt.Sprintf("%d", snap.Messages), true),
		utils.Field("Commandes exécutées", fmt.Sprintf("%d", snap.TotalCommands), true),
		utils.Field("Timeouts anti-spam", fmt.Sprintf("%d", snap.SpamTimeouts), true),
		utils.Field("Erreurs DB", fmt.Sprintf("%d", snap.DBErrors), true),
		utils.Field("Top commandes", topCmds, false),
	)
	utils.RespondEmbed(s, i.Interaction, embed)
}
