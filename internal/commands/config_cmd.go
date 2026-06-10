package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/utils"
)

func (h *Handler) cmdSetLevelUpChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Permission refusée."))
		return
	}
	ch := i.ApplicationCommandData().Options[0].ChannelValue(s)
	if err := h.db.SetLevelUpChannelID(i.GuildID, ch.ID); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Erreur lors de la sauvegarde."))
		return
	}
	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("✅ Canal level-up configuré",
			fmt.Sprintf("Les annonces de level-up et de prestige seront envoyées dans <#%s>.", ch.ID),
			utils.ColorGreen))
}

func (h *Handler) cmdSetDailyCooldown(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Permission refusée."))
		return
	}
	hours := int(i.ApplicationCommandData().Options[0].IntValue())
	if hours < 1 || hours > 168 {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("La valeur doit être entre **1** et **168** heures."))
		return
	}
	if err := h.db.SetDailyCooldownHours(i.GuildID, hours); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Erreur lors de la sauvegarde."))
		return
	}
	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("✅ Cooldown /daily configuré",
			fmt.Sprintf("Le cooldown de `/daily` est maintenant de **%dh**.", hours),
			utils.ColorGreen))
}

func (h *Handler) cmdSetWorkCooldown(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Permission refusée."))
		return
	}
	hours := int(i.ApplicationCommandData().Options[0].IntValue())
	if hours < 1 || hours > 168 {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("La valeur doit être entre **1** et **168** heures."))
		return
	}
	if err := h.db.SetWorkCooldownHours(i.GuildID, hours); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Erreur lors de la sauvegarde."))
		return
	}
	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("✅ Cooldown /work configuré",
			fmt.Sprintf("Le cooldown de `/work` est maintenant de **%dh**.", hours),
			utils.ColorGreen))
}

func (h *Handler) cmdSetMaxBet(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Permission refusée."))
		return
	}
	amount := i.ApplicationCommandData().Options[0].IntValue()
	if amount < 0 {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Le montant doit être **≥ 0** (0 = illimité)."))
		return
	}
	if err := h.db.SetMaxBet(i.GuildID, amount); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Erreur lors de la sauvegarde."))
		return
	}
	var msg string
	if amount == 0 {
		msg = "La mise maximale est désormais **illimitée**."
	} else {
		msg = fmt.Sprintf("La mise maximale est désormais de **%s** 🪙.", coins(amount))
	}
	utils.RespondEmbed(s, i.Interaction, utils.Embed("✅ Mise maximale configurée", msg, utils.ColorGreen))
}

func (h *Handler) cmdConfig(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Permission refusée."))
		return
	}
	cfg, err := h.db.GetGuildConfig(i.GuildID)
	if err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Erreur lecture configuration."))
		return
	}

	chanField := func(id string) string {
		if id != "" {
			return fmt.Sprintf("<#%s>", id)
		}
		return "Non configuré"
	}
	roleField := func(id string) string {
		if id != "" {
			return fmt.Sprintf("<@&%s>", id)
		}
		return "Non configuré"
	}

	dailyHrs := cfg.DailyCooldownHours
	if dailyHrs == 0 {
		dailyHrs = 24
	}
	workHrs := cfg.WorkCooldownHours
	if workHrs == 0 {
		workHrs = 4
	}

	maxBetStr := "Illimité"
	if cfg.MaxBet > 0 {
		maxBetStr = fmt.Sprintf("**%s** 🪙", coins(cfg.MaxBet))
	}

	embed := utils.EmbedFields(
		"⚙️ Configuration du serveur", "", utils.ColorBlue,
		utils.Field("📢 Canal level-up", chanField(cfg.LevelUpChannelID), true),
		utils.Field("📋 Canal logs", chanField(cfg.LogChannelID), true),
		utils.Field("🎭 Auto-rôle", roleField(cfg.AutoRoleID), true),
		utils.Field("🎫 Rôle staff", roleField(cfg.StaffRoleID), true),
		utils.Field("📌 Canal rôles", chanField(cfg.RolesChannelID), true),
		utils.Field("🎫 Catégorie tickets", chanField(cfg.TicketCategoryID), true),
		utils.Field("⏱️ Cooldown /daily", fmt.Sprintf("**%dh**", dailyHrs), true),
		utils.Field("⏱️ Cooldown /work", fmt.Sprintf("**%dh**", workHrs), true),
		utils.Field("🎰 Mise maximale", maxBetStr, true),
	)
	utils.RespondEmbedEphemeral(s, i.Interaction, embed)
}
