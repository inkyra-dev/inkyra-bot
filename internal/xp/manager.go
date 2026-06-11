package xp

import (
	"fmt"
	"log/slog"
	"math/rand"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/achievements"
	"discord-bot/internal/metrics"
	"discord-bot/internal/repositories"
	"discord-bot/internal/utils"
)

const (
	minXPGain       = 15
	maxXPGain       = 25
	messageCooldown = 60 // seconds
)

// ChannelSettings est implémenté par *database.DB pour lire levelup_channel_id.
type ChannelSettings interface {
	GetLevelUpChannelID(guildID string) (string, error)
}

type Manager struct {
	repo      *repositories.XPRepo
	statsRepo *repositories.StatsRepo
	achMgr    *achievements.Manager
	settings  ChannelSettings
}

func NewManager(repo *repositories.XPRepo, statsRepo *repositories.StatsRepo, achMgr *achievements.Manager, settings ChannelSettings) *Manager {
	return &Manager{repo: repo, statsRepo: statsRepo, achMgr: achMgr, settings: settings}
}

var levelAchievements = map[int]string{
	5: "level_5", 10: "level_10", 25: "level_25", 50: "level_50", 100: "level_100",
}

// HandleMessage est appelé à chaque message d'utilisateur pour attribuer de l'XP.
func (m *Manager) HandleMessage(s *discordgo.Session, guildID, userID, channelID string) {
	metrics.GetMetrics().IncrMessage()

	// Compteur brut de messages (sans cooldown) pour les achievements.
	if n, err := m.statsRepo.IncrementMessages(guildID, userID); err == nil {
		switch n {
		case 1:
			m.achMgr.Check(guildID, userID, "first_message")
		case 100:
			m.achMgr.Check(guildID, userID, "msg_100")
		case 500:
			m.achMgr.Check(guildID, userID, "msg_500")
		case 1000:
			m.achMgr.Check(guildID, userID, "msg_1000")
		}
	}

	ok, err := m.repo.CheckAndSetCooldown(guildID, userID, messageCooldown)
	if err != nil || !ok {
		return
	}

	prev, err := m.repo.Get(guildID, userID)
	if err != nil {
		return
	}
	oldLevel := LevelFromXP(prev.TotalXP)

	gain := int64(minXPGain + rand.Intn(maxXPGain-minXPGain+1))
	newTotal, err := m.repo.AddXP(guildID, userID, gain)
	if err != nil {
		slog.Error("AddXP échoué", "component", "xp", "guild_id", guildID, "user_id", userID, "error", err)
		metrics.GetMetrics().IncrDBError()
		return
	}

	newLevel := LevelFromXP(newTotal)
	if newLevel != oldLevel {
		if err := m.repo.UpdateLevel(guildID, userID, newLevel); err != nil {
			slog.Error("UpdateLevel échoué", "component", "xp", "guild_id", guildID, "user_id", userID, "error", err)
			metrics.GetMetrics().IncrDBError()
		}
		if key, ok := levelAchievements[newLevel]; ok {
			m.achMgr.Check(guildID, userID, key)
		}
		m.announceLevelUp(s, guildID, userID, channelID, newLevel, prev.Prestige)
	}
}

func (m *Manager) announceLevelUp(s *discordgo.Session, guildID, userID, fallbackChannelID string, level, prestige int) {
	target := fallbackChannelID
	if m.settings != nil {
		if ch, err := m.settings.GetLevelUpChannelID(guildID); err == nil && ch != "" {
			target = ch
		}
	}

	desc := fmt.Sprintf("<@%s> a atteint le **niveau %d** ! 🎉", userID, level)
	embed := utils.EmbedFields("⬆️ Level Up !", desc, utils.ColorPurple,
		utils.Field("Niveau", fmt.Sprintf("**%d**", level), true),
	)
	if prestige > 0 {
		embed.Fields = append(embed.Fields, utils.Field("Prestige", fmt.Sprintf("⭐ %d", prestige), true))
	}
	s.ChannelMessageSendEmbed(target, embed)
}

// Get retourne (ou crée) les données XP d'un utilisateur.
func (m *Manager) Get(guildID, userID string) (*repositories.UserXP, error) {
	return m.repo.GetOrCreate(guildID, userID)
}

// GetLeaderboard retourne la page N du leaderboard (page commence à 1).
func (m *Manager) GetLeaderboard(guildID string, page int) ([]repositories.LeaderboardEntry, int64, error) {
	const pageSize = 10
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	entries, err := m.repo.GetLeaderboard(guildID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := m.repo.Count(guildID)
	return entries, total, err
}

// GetRank retourne le rang et le nombre total d'utilisateurs.
func (m *Manager) GetRank(guildID, userID string) (rank, total int64, err error) {
	return m.repo.GetRank(guildID, userID)
}

func (m *Manager) AddXP(guildID, userID string, amount int64) error {
	_, err := m.repo.AddXP(guildID, userID, amount)
	return err
}

func (m *Manager) RemoveXP(guildID, userID string, amount int64) error {
	data, err := m.repo.Get(guildID, userID)
	if err != nil {
		return err
	}
	newXP := data.TotalXP - amount
	if newXP < 0 {
		newXP = 0
	}
	return m.repo.SetXP(guildID, userID, newXP, LevelFromXP(newXP))
}

func (m *Manager) SetXP(guildID, userID string, amount int64) error {
	return m.repo.SetXP(guildID, userID, amount, LevelFromXP(amount))
}

func (m *Manager) SetLevel(guildID, userID string, level int) error {
	totalXP := TotalXPForLevel(level)
	return m.repo.SetXP(guildID, userID, totalXP, level)
}

func (m *Manager) Prestige(guildID, userID string) error {
	data, err := m.repo.Get(guildID, userID)
	if err != nil {
		return err
	}
	if data.Level < MaxLevel {
		return fmt.Errorf("tu dois être niveau %d pour prestigier (niveau actuel: %d)", MaxLevel, data.Level)
	}
	if err := m.repo.Prestige(guildID, userID); err != nil {
		return err
	}
	m.achMgr.Check(guildID, userID, "prestige")
	return nil
}

func (m *Manager) Reset(guildID, userID string) error {
	return m.repo.Reset(guildID, userID)
}
