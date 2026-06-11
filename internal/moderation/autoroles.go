package moderation

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/database"
	"discord-bot/internal/utils"
)

type RolesManager struct {
	db *database.DB
}

func NewRolesManager(db *database.DB) *RolesManager {
	return &RolesManager{db: db}
}

func (m *RolesManager) HandleMemberJoin(s *discordgo.Session, guildID, userID string) {
	roleID, err := m.db.GetAutoRoleID(guildID)
	if err != nil || roleID == "" {
		return
	}
	if err := s.GuildMemberRoleAdd(guildID, userID, roleID); err != nil {
		slog.Error("attribution rôle échouée", "component", "autoroles", "role_id", roleID, "user_id", userID, "error", err)
	}
}

func (m *RolesManager) SetAutoRole(guildID, roleID string) error {
	return m.db.SetAutoRoleID(guildID, roleID)
}

// SetupRoles poste l'embed de sélection de rôles dans channelID, l'épingle, et enregistre l'info.
func (m *RolesManager) SetupRoles(s *discordgo.Session, guildID, channelID string) error {
	buttons, err := m.db.GetRoleButtons(guildID)
	if err != nil {
		return err
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🎭 Sélection de rôles",
		Description: "Clique sur un bouton pour obtenir ou retirer un rôle.",
		Color:       utils.ColorPurple,
	}

	msg, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: buildRoleComponents(buttons),
	})
	if err != nil {
		return fmt.Errorf("envoi embed: %w", err)
	}

	if err := s.ChannelMessagePin(channelID, msg.ID); err != nil {
		slog.Warn("épinglage échoué", "component", "autoroles", "error", err)
	}

	return m.db.SetRolesChannelAndMessage(guildID, channelID, msg.ID)
}

// AddRoleButton ajoute un bouton en DB et rafraîchit l'embed existant si présent.
func (m *RolesManager) AddRoleButton(s *discordgo.Session, guildID, roleID, label, emoji string) error {
	if err := m.db.AddRoleButton(guildID, roleID, label, emoji); err != nil {
		return err
	}
	return m.refreshEmbed(s, guildID)
}

func (m *RolesManager) refreshEmbed(s *discordgo.Session, guildID string) error {
	channelID, messageID, err := m.db.GetRolesMessageInfo(guildID)
	if err != nil || channelID == "" || messageID == "" {
		return nil
	}

	buttons, err := m.db.GetRoleButtons(guildID)
	if err != nil {
		return err
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🎭 Sélection de rôles",
		Description: "Clique sur un bouton pour obtenir ou retirer un rôle.",
		Color:       utils.ColorPurple,
	}

	embeds := []*discordgo.MessageEmbed{embed}
	comps := buildRoleComponents(buttons)
	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         messageID,
		Channel:    channelID,
		Embeds:     &embeds,
		Components: &comps,
	})
	return err
}

// HandleRoleToggle gère le clic sur un bouton de rôle (ajoute ou retire le rôle).
func (m *RolesManager) HandleRoleToggle(s *discordgo.Session, i *discordgo.InteractionCreate, roleID string) {
	userID := i.Member.User.ID
	guildID := i.GuildID

	hasRole := false
	for _, r := range i.Member.Roles {
		if r == roleID {
			hasRole = true
			break
		}
	}

	var err error
	var msg string
	if hasRole {
		err = s.GuildMemberRoleRemove(guildID, userID, roleID)
		msg = fmt.Sprintf("✅ Rôle <@&%s> retiré.", roleID)
	} else {
		err = s.GuildMemberRoleAdd(guildID, userID, roleID)
		msg = fmt.Sprintf("✅ Rôle <@&%s> attribué.", roleID)
	}

	if err != nil {
		utils.RespondEphemeral(s, i.Interaction, "❌ Impossible de modifier le rôle. Vérifie les permissions du bot.")
		return
	}
	utils.RespondEphemeral(s, i.Interaction, msg)
}

func buildRoleComponents(buttons []database.RoleButton) []discordgo.MessageComponent {
	if len(buttons) == 0 {
		return nil
	}

	var rows []discordgo.MessageComponent
	var rowBtns []discordgo.MessageComponent

	for i, b := range buttons {
		btn := discordgo.Button{
			CustomID: "role_toggle:" + b.RoleID,
			Label:    b.Label,
			Style:    discordgo.PrimaryButton,
		}
		if b.Emoji != "" {
			btn.Emoji = &discordgo.ComponentEmoji{Name: b.Emoji}
		}
		rowBtns = append(rowBtns, btn)

		if len(rowBtns) == 5 || i == len(buttons)-1 {
			rows = append(rows, discordgo.ActionsRow{Components: rowBtns})
			rowBtns = nil
		}
	}

	return rows
}
