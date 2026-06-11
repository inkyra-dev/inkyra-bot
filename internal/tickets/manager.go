package tickets

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/config"
	"discord-bot/internal/database"
)

var (
	ErrAlreadyOpen = errors.New("tu as déjà un ticket ouvert")
	ErrNotATicket  = errors.New("ce salon n'est pas un ticket")
)

type Manager struct {
	db  *database.DB
	cfg *config.Config
}

func NewManager(db *database.DB, cfg *config.Config) *Manager {
	return &Manager{db: db, cfg: cfg}
}

func (m *Manager) Open(s *discordgo.Session, guildID, userID, username string) (*discordgo.Channel, error) {
	existing, err := m.db.GetOpenTicketByUser(guildID, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyOpen
	}

	data := discordgo.GuildChannelCreateData{
		Name:     fmt.Sprintf("ticket-%s", username),
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: m.cfg.TicketCategoryID,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			// Interdire @everyone
			{ID: guildID, Type: discordgo.PermissionOverwriteTypeRole, Deny: discordgo.PermissionViewChannel},
			// Permettre au bot
			{
				ID:   s.State.User.ID,
				Type: discordgo.PermissionOverwriteTypeMember,
				Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages |
					discordgo.PermissionReadMessageHistory | discordgo.PermissionManageChannels,
			},
			// Permettre à l'auteur
			{
				ID:   userID,
				Type: discordgo.PermissionOverwriteTypeMember,
				Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages |
					discordgo.PermissionReadMessageHistory,
			},
		},
	}

	if m.cfg.StaffRoleID != "" {
		data.PermissionOverwrites = append(data.PermissionOverwrites, &discordgo.PermissionOverwrite{
			ID:   m.cfg.StaffRoleID,
			Type: discordgo.PermissionOverwriteTypeRole,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages |
				discordgo.PermissionReadMessageHistory,
		})
	}

	ch, err := s.GuildChannelCreateComplex(guildID, data)
	if err != nil {
		return nil, fmt.Errorf("création salon: %w", err)
	}

	if err := m.db.CreateTicket(ch.ID, guildID, userID); err != nil {
		s.ChannelDelete(ch.ID)
		return nil, fmt.Errorf("insertion DB: %w", err)
	}

	return ch, nil
}

func (m *Manager) Close(s *discordgo.Session, channelID, closedByID string) error {
	ticket, err := m.db.GetTicketByChannel(channelID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotATicket
		}
		return err
	}

	if err := m.db.CloseTicket(channelID); err != nil {
		return err
	}

	if m.cfg.LogChannelID != "" {
		transcript := NewTranscript(s, channelID)
		content, _ := transcript.Generate()
		m.logClosure(s, ticket, closedByID, content)
	}

	slog.Info("ticket fermé", "component", "tickets", "channel_id", channelID, "closed_by", closedByID)

	time.AfterFunc(5*time.Second, func() {
		if _, err := s.ChannelDelete(channelID); err != nil {
			slog.Error("suppression salon échouée", "component", "tickets", "channel_id", channelID, "error", err)
		}
	})

	return nil
}

func (m *Manager) AddUser(s *discordgo.Session, channelID, userID string) error {
	if !m.db.IsOpenTicket(channelID) {
		return ErrNotATicket
	}
	return s.ChannelPermissionSet(channelID, userID, discordgo.PermissionOverwriteTypeMember,
		discordgo.PermissionViewChannel|discordgo.PermissionSendMessages|discordgo.PermissionReadMessageHistory, 0,
	)
}

func (m *Manager) RemoveUser(s *discordgo.Session, channelID, userID string) error {
	if !m.db.IsOpenTicket(channelID) {
		return ErrNotATicket
	}
	return s.ChannelPermissionSet(channelID, userID, discordgo.PermissionOverwriteTypeMember,
		0, discordgo.PermissionViewChannel|discordgo.PermissionSendMessages|discordgo.PermissionReadMessageHistory,
	)
}

func (m *Manager) logClosure(s *discordgo.Session, t *database.Ticket, closedBy, transcript string) {
	msg := fmt.Sprintf(
		"**Ticket fermé** `#%s`\n**Ouvert par :** <@%s>\n**Fermé par :** <@%s>",
		t.ChannelID, t.UserID, closedBy,
	)
	s.ChannelMessageSend(m.cfg.LogChannelID, msg)

	if transcript != "" {
		// Envoi du transcript en fichier pour éviter la limite des 2000 caractères
		s.ChannelFileSend(m.cfg.LogChannelID, "transcript.txt",
			strings.NewReader(fmt.Sprintf("Transcript du ticket %s\n\n%s", t.ChannelID, transcript)))
	}
}
