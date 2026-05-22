package commands

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/tickets"
	"discord-bot/internal/utils"
)

func (h *Handler) cmdTicket(s *discordgo.Session, i *discordgo.InteractionCreate) {
	member := i.Member
	if member == nil {
		utils.RespondEphemeral(s, i.Interaction, "Cette commande ne fonctionne qu'en serveur.")
		return
	}

	ch, err := h.tickets.Open(s, i.GuildID, member.User.ID, member.User.Username)
	if err != nil {
		if errors.Is(err, tickets.ErrAlreadyOpen) {
			utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Tu as déjà un ticket ouvert."))
			return
		}
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Impossible de créer le ticket : "+err.Error()))
		return
	}

	// Confirmer à l'auteur
	utils.RespondEmbedEphemeral(s, i.Interaction,
		utils.Embed("Ticket ouvert", fmt.Sprintf("Ton ticket a été créé : <#%s>", ch.ID), utils.ColorGreen),
	)

	// Embed de bienvenue dans le salon du ticket avec boutons
	welcomeEmbed := utils.EmbedFields(
		"Ticket de support",
		fmt.Sprintf("Bienvenue <@%s> !\nExplique ton problème, le staff va te répondre.", member.User.ID),
		utils.ColorBlue,
		utils.Field("Auteur", fmt.Sprintf("<@%s>", member.User.ID), true),
	)

	buttons := []discordgo.MessageComponent{
		discordgo.ActionsRow{Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Fermer le ticket",
				Style:    discordgo.DangerButton,
				CustomID: "close_ticket",
				Emoji:    &discordgo.ComponentEmoji{Name: "🔒"},
			},
			discordgo.Button{
				Label:    "Transcription",
				Style:    discordgo.SecondaryButton,
				CustomID: "transcript_ticket",
				Emoji:    &discordgo.ComponentEmoji{Name: "📄"},
			},
		}},
	}

	s.ChannelMessageSendComplex(ch.ID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{welcomeEmbed},
		Components: buttons,
	})
}

func (h *Handler) cmdClose(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if err := h.closeChannel(s, i.Interaction, i.ChannelID, i.Member.User.ID); err != nil {
		return
	}
}

func (h *Handler) closeTicketComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h.closeChannel(s, i.Interaction, i.ChannelID, i.Member.User.ID)
}

func (h *Handler) transcriptTicketComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	utils.RespondEphemeral(s, i.Interaction, "Génération de la transcription en cours...")
	// Le transcript complet est envoyé dans les logs lors de la fermeture
	utils.RespondEphemeral(s, i.Interaction, "Utilise /close pour fermer le ticket et générer le transcript automatiquement.")
}

func (h *Handler) closeChannel(s *discordgo.Session, interaction *discordgo.Interaction, channelID, userID string) error {
	utils.Respond(s, interaction, "🔒 Fermeture du ticket dans 5 secondes...")

	if err := h.tickets.Close(s, channelID, userID); err != nil {
		if errors.Is(err, tickets.ErrNotATicket) {
			utils.RespondEphemeral(s, interaction, "Ce salon n'est pas un ticket.")
			return err
		}
		utils.RespondEphemeral(s, interaction, "Erreur lors de la fermeture : "+err.Error())
		return err
	}
	return nil
}

func (h *Handler) cmdAddUser(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return
	}
	targetUser := options[0].UserValue(s)

	if err := h.tickets.AddUser(s, i.ChannelID, targetUser.ID); err != nil {
		if errors.Is(err, tickets.ErrNotATicket) {
			utils.RespondEphemeral(s, i.Interaction, "Ce salon n'est pas un ticket.")
			return
		}
		utils.RespondEphemeral(s, i.Interaction, "Erreur : "+err.Error())
		return
	}

	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("Utilisateur ajouté", fmt.Sprintf("<@%s> a accès à ce ticket.", targetUser.ID), utils.ColorGreen),
	)
}

func (h *Handler) cmdRemoveUser(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return
	}
	targetUser := options[0].UserValue(s)

	if err := h.tickets.RemoveUser(s, i.ChannelID, targetUser.ID); err != nil {
		if errors.Is(err, tickets.ErrNotATicket) {
			utils.RespondEphemeral(s, i.Interaction, "Ce salon n'est pas un ticket.")
			return
		}
		utils.RespondEphemeral(s, i.Interaction, "Erreur : "+err.Error())
		return
	}

	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("Utilisateur retiré", fmt.Sprintf("<@%s> n'a plus accès à ce ticket.", targetUser.ID), utils.ColorYellow),
	)
}
