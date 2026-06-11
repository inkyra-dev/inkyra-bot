package moderation

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/utils"
)

type Logger struct {
	logChanID string
}

func NewLogger(logChanID string) *Logger {
	return &Logger{logChanID: logChanID}
}

func (l *Logger) send(s *discordgo.Session, embed *discordgo.MessageEmbed) {
	if l.logChanID == "" {
		return
	}
	if _, err := s.ChannelMessageSendEmbed(l.logChanID, embed); err != nil {
		slog.Error("erreur envoi log", "component", "modlog", "error", err)
	}
}

func (l *Logger) LogTimeout(s *discordgo.Session, userID, username, channelID string) {
	l.send(s, &discordgo.MessageEmbed{
		Title: "🔇 Anti-spam — Timeout",
		Color: utils.ColorYellow,
		Fields: []*discordgo.MessageEmbedField{
			utils.Field("Utilisateur", fmt.Sprintf("<@%s> (`%s`)", userID, username), true),
			utils.Field("Canal", fmt.Sprintf("<#%s>", channelID), true),
			utils.Field("Durée", "5 minutes", true),
			utils.Field("Raison", "Spam (>5 messages en 5s)", false),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func (l *Logger) LogBan(s *discordgo.Session, guildID, targetID string) {
	if l.logChanID == "" {
		return
	}
	go func() {
		time.Sleep(600 * time.Millisecond)
		modID := fetchModerator(s, guildID, targetID, discordgo.AuditLogActionMemberBanAdd)
		modStr := "Inconnu"
		if modID != "" {
			modStr = fmt.Sprintf("<@%s>", modID)
		}
		l.send(s, &discordgo.MessageEmbed{
			Title: "🔨 Membre banni",
			Color: utils.ColorRed,
			Fields: []*discordgo.MessageEmbedField{
				utils.Field("Utilisateur", fmt.Sprintf("<@%s>", targetID), true),
				utils.Field("Modérateur", modStr, true),
			},
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}()
}

func (l *Logger) LogUnban(s *discordgo.Session, guildID, targetID string) {
	if l.logChanID == "" {
		return
	}
	go func() {
		time.Sleep(600 * time.Millisecond)
		modID := fetchModerator(s, guildID, targetID, discordgo.AuditLogActionMemberBanRemove)
		modStr := "Inconnu"
		if modID != "" {
			modStr = fmt.Sprintf("<@%s>", modID)
		}
		l.send(s, &discordgo.MessageEmbed{
			Title: "✅ Membre débanni",
			Color: utils.ColorGreen,
			Fields: []*discordgo.MessageEmbedField{
				utils.Field("Utilisateur", fmt.Sprintf("<@%s>", targetID), true),
				utils.Field("Modérateur", modStr, true),
			},
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}()
}

func fetchModerator(s *discordgo.Session, guildID, targetID string, actionType discordgo.AuditLogAction) string {
	logs, err := s.GuildAuditLog(guildID, "", "", int(actionType), 5)
	if err != nil {
		return ""
	}
	for _, entry := range logs.AuditLogEntries {
		if entry.TargetID == targetID {
			return entry.UserID
		}
	}
	return ""
}
