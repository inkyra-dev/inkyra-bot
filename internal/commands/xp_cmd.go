package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/utils"
	"discord-bot/internal/xp"
)

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
	page := 1
	if opts := i.ApplicationCommandData().Options; len(opts) > 0 {
		page = int(opts[0].IntValue())
		if page < 1 {
			page = 1
		}
	}

	entries, total, err := h.xpMgr.GetLeaderboard(i.GuildID, page)
	if err != nil || len(entries) == 0 {
		utils.RespondEphemeral(s, i.Interaction, "Aucune donnée XP pour ce serveur.")
		return
	}

	const pageSize = 10
	totalPages := (int(total) + pageSize - 1) / pageSize

	medals := []string{"🥇", "🥈", "🥉"}
	var sb strings.Builder
	for _, e := range entries {
		medal := ""
		rank := int(e.Rank)
		if rank-1 < len(medals) {
			medal = medals[rank-1] + " "
		}
		name := usernameOrID(s, i.GuildID, e.UserID)
		sb.WriteString(fmt.Sprintf("%s`#%d` **%s** — Niv. %d | %d XP\n",
			medal, rank, name, e.Level, e.TotalXP))
	}

	embed := utils.EmbedFields(
		"🏆 Classement XP",
		sb.String(),
		utils.ColorPurple,
		utils.Field("Page", fmt.Sprintf("%d / %d", page, totalPages), true),
	)

	buttons := leaderboardButtons(page, totalPages, "xplb")
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: buttons,
		},
	})
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
	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("⭐ Prestige !", fmt.Sprintf("<@%s> a prestigié ! XP remis à zéro.", i.Member.User.ID), utils.ColorPurple))
}
