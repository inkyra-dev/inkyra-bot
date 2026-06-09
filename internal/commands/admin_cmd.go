package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/utils"
)

// isAdmin vérifie si l'utilisateur a la permission Administrateur sur le serveur.
func isAdmin(s *discordgo.Session, guildID, userID string) bool {
	member, err := s.State.Member(guildID, userID)
	if err != nil {
		member, err = s.GuildMember(guildID, userID)
		if err != nil {
			return false
		}
	}
	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			continue
		}
		if role.Permissions&discordgo.PermissionAdministrator != 0 {
			return true
		}
	}
	return false
}

func (h *Handler) cmdGiveMoney(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	if err := h.econMgr.AdminGive(i.GuildID, target.ID, amount); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}
	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("💰 Argent ajouté",
			fmt.Sprintf("**+%s** 🪙 ajoutés au portefeuille de <@%s>", coins(amount), target.ID),
			utils.ColorGreen))
}

func (h *Handler) cmdRemoveMoney(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	if err := h.econMgr.AdminRemove(i.GuildID, target.ID, amount); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}
	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("💸 Argent retiré",
			fmt.Sprintf("**-%s** 🪙 retirés du portefeuille de <@%s>", coins(amount), target.ID),
			utils.ColorYellow))
}

func (h *Handler) cmdResetUser(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Permission refusée."))
		return
	}
	opts := i.ApplicationCommandData().Options
	target := opts[0].UserValue(s)

	if err := h.xpMgr.Reset(i.GuildID, target.ID); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(fmt.Sprintf("Erreur réinitialisation XP : %v", err)))
		return
	}
	if err := h.econMgr.AdminReset(i.GuildID, target.ID); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(fmt.Sprintf("XP réinitialisé, mais erreur économie : %v", err)))
		return
	}

	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("🔄 Réinitialisation",
			fmt.Sprintf("XP et économie de <@%s> remis à zéro.", target.ID),
			utils.ColorRed))
}

func (h *Handler) cmdSetXP(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Permission refusée."))
		return
	}
	opts := i.ApplicationCommandData().Options
	target := opts[0].UserValue(s)
	amount := opts[1].IntValue()
	if amount < 0 {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Le montant d'XP ne peut pas être négatif."))
		return
	}

	if err := h.xpMgr.SetXP(i.GuildID, target.ID, amount); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}
	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("✨ XP définie",
			fmt.Sprintf("<@%s> a maintenant **%d XP**.", target.ID, amount),
			utils.ColorGreen))
}

func (h *Handler) cmdEconomyReset(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Permission refusée."))
		return
	}
	opts := i.ApplicationCommandData().Options
	target := opts[0].UserValue(s)

	if err := h.econMgr.AdminReset(i.GuildID, target.ID); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}
	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("🔄 Économie réinitialisée",
			fmt.Sprintf("Compte de <@%s> remis à zéro (portefeuille=0, banque=100).", target.ID),
			utils.ColorRed))
}
