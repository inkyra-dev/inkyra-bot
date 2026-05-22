package xp

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/repositories"
	"discord-bot/internal/utils"
)

const (
	minXPGain       = 15
	maxXPGain       = 25
	messageCooldown = 60 * time.Second
)

type Manager struct {
	mu        sync.RWMutex
	cooldowns map[string]time.Time // key: guildID+":"+userID
	repo      *repositories.XPRepo
}

func NewManager(repo *repositories.XPRepo) *Manager {
	m := &Manager{
		cooldowns: make(map[string]time.Time),
		repo:      repo,
	}
	go m.cleanupLoop()
	return m
}

func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		m.mu.Lock()
		for k, v := range m.cooldowns {
			if now.After(v) {
				delete(m.cooldowns, k)
			}
		}
		m.mu.Unlock()
	}
}

// HandleMessage est appelé à chaque message d'utilisateur pour attribuer de l'XP.
func (m *Manager) HandleMessage(s *discordgo.Session, guildID, userID, channelID string) {
	key := guildID + ":" + userID

	m.mu.RLock()
	expiry, ok := m.cooldowns[key]
	m.mu.RUnlock()

	if ok && time.Now().Before(expiry) {
		return
	}

	m.mu.Lock()
	m.cooldowns[key] = time.Now().Add(messageCooldown)
	m.mu.Unlock()

	prev, err := m.repo.Get(guildID, userID)
	if err != nil {
		return
	}
	oldLevel := LevelFromXP(prev.TotalXP)

	gain := int64(minXPGain + rand.Intn(maxXPGain-minXPGain+1))
	newTotal, err := m.repo.AddXP(guildID, userID, gain)
	if err != nil {
		log.Printf("[xp] AddXP error: %v", err)
		return
	}

	newLevel := LevelFromXP(newTotal)
	if newLevel != oldLevel {
		if err := m.repo.UpdateLevel(guildID, userID, newLevel); err != nil {
			log.Printf("[xp] UpdateLevel error: %v", err)
		}
		m.announceLevelUp(s, guildID, userID, channelID, newLevel, prev.Prestige)
	}
}

func (m *Manager) announceLevelUp(s *discordgo.Session, guildID, userID, channelID string, level, prestige int) {
	desc := fmt.Sprintf("<@%s> a atteint le **niveau %d** ! 🎉", userID, level)
	embed := utils.EmbedFields("⬆️ Level Up !", desc, utils.ColorPurple,
		utils.Field("Niveau", fmt.Sprintf("**%d**", level), true),
	)
	if prestige > 0 {
		embed.Fields = append(embed.Fields, utils.Field("Prestige", fmt.Sprintf("⭐ %d", prestige), true))
	}
	s.ChannelMessageSendEmbed(channelID, embed)
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
	return m.repo.Prestige(guildID, userID)
}

func (m *Manager) Reset(guildID, userID string) error {
	return m.repo.Reset(guildID, userID)
}
