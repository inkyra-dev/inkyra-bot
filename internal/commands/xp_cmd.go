package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/repositories"
	"discord-bot/internal/utils"
	"discord-bot/internal/xp"
)

func xpLeaderboardBody(s *discordgo.Session, guildID string, entries []repositories.LeaderboardEntry) string {
	medals := []string{"🥇", "🥈", "🥉"}
	var sb strings.Builder
	for _, e := range entries {
		medal := ""
		if int(e.Rank)-1 < len(medals) {
			medal = medals[e.Rank-1] + " "
		}
		name := usernameOrID(s, guildID, e.UserID)
		fmt.Fprintf(&sb, "%s`#%d` **%s** — Niv. %d | %d XP\n", medal, e.Rank, name, e.Level, e.TotalXP)
	}
	return sb.String()
}

func (h *Handler) cmdRank(s *discordgo.Session, i *discordgo.InteractionCreate) {
	targetID := i.Member.User.ID
	targetName := i.Member.User.Username

	if opts := i.ApplicationCommandData().Options; len(opts) > 0 {
		u := opts[0].UserValue(s)
		targetID = u.ID
		targetName = u.Username
	}

	data, err := h.xpMgr.Get(i.GuildID, targetID)
	if err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Erreur récupération XP."))
		return
	}

	rank, total, _ := h.xpMgr.GetRank(i.GuildID, targetID)
	level := xp.LevelFromXP(data.TotalXP)
	cur, needed := xp.ProgressInLevel(data.TotalXP, level)
	bar := xp.ProgressBar(cur, needed, 12)

	embed := utils.EmbedFields(
		fmt.Sprintf("Profil XP — %s", targetName),
		fmt.Sprintf("`%s` **%d / %d XP**", bar, cur, needed),
		utils.ColorPurple,
		utils.Field("🏅 Rang", fmt.Sprintf("#%d / %d", rank, total), true),
		utils.Field("⬆️ Niveau", fmt.Sprintf("**%d**", level), true),
		utils.Field("✨ XP Total", fmt.Sprintf("%d", data.TotalXP), true),
	)
	if data.Prestige > 0 {
		embed.Fields = append(embed.Fields, utils.Field("⭐ Prestige", fmt.Sprintf("%d", data.Prestige), true))
	}
	utils.RespondEmbed(s, i.Interaction, embed)
}

func (h *Handler) cmdLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	leaderType := "xp"
	page := 1
	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case "type":
			leaderType = opt.StringValue()
		case "page":
			if v := int(opt.IntValue()); v >= 1 {
				page = v
			}
		}
	}

	if leaderType == "economie" {
		h.showEconLeaderboard(s, i, page, false)
	} else {
		h.showXPLeaderboard(s, i, page, false)
	}
}

func (h *Handler) showXPLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate, page int, update bool) {
	entries, total, err := h.xpMgr.GetLeaderboard(i.GuildID, page)
	if err != nil || len(entries) == 0 {
		utils.RespondEphemeral(s, i.Interaction, "Aucune donnée XP pour ce serveur.")
		return
	}
	const pageSize = 10
	totalPages := (int(total) + pageSize - 1) / pageSize
	respondLeaderboard(s, i, "🏆 Classement XP", xpLeaderboardBody(s, i.GuildID, entries), utils.ColorPurple, page, totalPages, "xplb", update)
}

func (h *Handler) showEconLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate, page int, update bool) {
	const pageSize = 10
	offset := (page - 1) * pageSize
	entries, err := h.econMgr.GetLeaderboard(i.GuildID, pageSize, offset)
	if err != nil || len(entries) == 0 {
		utils.RespondEphemeral(s, i.Interaction, "Aucune donnée économique pour ce serveur.")
		return
	}
	total, _ := h.econMgr.Count(i.GuildID)
	totalPages := (int(total) + pageSize - 1) / pageSize
	respondLeaderboard(s, i, "💰 Classement Économie", econLeaderboardBody(s, i.GuildID, entries, offset), utils.ColorYellow, page, totalPages, "ecolb", update)
}

// ── Admin XP ──────────────────────────────────────────────────────────────────

func (h *Handler) cmdAddXP(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Permission refusée."))
		return
	}
	opts := i.ApplicationCommandData().Options
	target := opts[0].UserValue(s)
	amount := opts[1].IntValue()
	if amount <= 0 {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Le montant doit être supérieur à 0."))
		return
	}

	if err := h.xpMgr.AddXP(i.GuildID, target.ID, amount); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}
	utils.RespondEmbed(s, i.Interaction, utils.Embed("XP ajoutée",
		fmt.Sprintf("**+%d XP** ajoutés à <@%s>", amount, target.ID), utils.ColorGreen))
}

func (h *Handler) cmdRemoveXP(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Permission refusée."))
		return
	}
	opts := i.ApplicationCommandData().Options
	target := opts[0].UserValue(s)
	amount := opts[1].IntValue()
	if amount <= 0 {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Le montant doit être supérieur à 0."))
		return
	}

	if err := h.xpMgr.RemoveXP(i.GuildID, target.ID, amount); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}
	utils.RespondEmbed(s, i.Interaction, utils.Embed("XP retirée",
		fmt.Sprintf("**-%d XP** retirés à <@%s>", amount, target.ID), utils.ColorYellow))
}

func (h *Handler) cmdSetLevel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Permission refusée."))
		return
	}
	opts := i.ApplicationCommandData().Options
	target := opts[0].UserValue(s)
	level := int(opts[1].IntValue())
	if level < 0 || level > xp.MaxLevel {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(fmt.Sprintf("Le niveau doit être entre 0 et %d.", xp.MaxLevel)))
		return
	}

	if err := h.xpMgr.SetLevel(i.GuildID, target.ID, level); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}
	utils.RespondEmbed(s, i.Interaction, utils.Embed("Niveau défini",
		fmt.Sprintf("<@%s> est maintenant niveau **%d**", target.ID, level), utils.ColorGreen))
}

func (h *Handler) cmdPrestige(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if err := h.xpMgr.Prestige(i.GuildID, i.Member.User.ID); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}
	embed := utils.Embed("⭐ Prestige !", fmt.Sprintf("<@%s> a prestigié ! XP remis à zéro.", i.Member.User.ID), utils.ColorPurple)
	utils.RespondEmbed(s, i.Interaction, embed)

	if ch, err := h.db.GetLevelUpChannelID(i.GuildID); err == nil && ch != "" && ch != i.ChannelID {
		s.ChannelMessageSendEmbed(ch, embed)
	}
}
