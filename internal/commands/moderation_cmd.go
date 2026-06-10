package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/utils"
)

func (h *Handler) cmdSetAutoRole(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Permission refusée."))
		return
	}
	role := i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID)
	if err := h.rolesMgr.SetAutoRole(i.GuildID, role.ID); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Erreur lors de la sauvegarde."))
		return
	}
	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("✅ Auto-rôle configuré",
			fmt.Sprintf("Les nouveaux membres recevront automatiquement le rôle <@&%s>.", role.ID),
			utils.ColorGreen))
}

func (h *Handler) cmdSetupRoles(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Permission refusée."))
		return
	}
	if err := h.rolesMgr.SetupRoles(s, i.GuildID, i.ChannelID); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(fmt.Sprintf("Erreur: %v", err)))
		return
	}
	utils.RespondEphemeral(s, i.Interaction, "✅ Embed de sélection de rôles posté et épinglé dans ce canal.")
}

func (h *Handler) cmdAddRoleButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i.GuildID, i.Member.User.ID) {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Permission refusée."))
		return
	}
	opts := i.ApplicationCommandData().Options
	role := opts[0].RoleValue(s, i.GuildID)
	label := opts[1].StringValue()
	emoji := ""
	if len(opts) > 2 {
		emoji = opts[2].StringValue()
	}

	if err := h.rolesMgr.AddRoleButton(s, i.GuildID, role.ID, label, emoji); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(fmt.Sprintf("Erreur: %v", err)))
		return
	}
	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("✅ Bouton ajouté",
			fmt.Sprintf("Bouton **%s** pour <@&%s> ajouté à l'embed de rôles.", label, role.ID),
			utils.ColorGreen))
}
