package moderation

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	spamLimit       = 5
	spamWindow      = 5 * time.Second
	timeoutDuration = 5 * time.Minute
)

type chanKey struct {
	guildID   string
	channelID string
	userID    string
}

type userKey struct {
	guildID string
	userID  string
}

type Monitor struct {
	mu     sync.Mutex
	counts map[chanKey]int
	warned map[userKey]bool
	logger *Logger
	done   chan struct{}
}

func NewMonitor(logger *Logger) *Monitor {
	m := &Monitor{
		counts: make(map[chanKey]int),
		warned: make(map[userKey]bool),
		logger: logger,
		done:   make(chan struct{}),
	}
	go m.resetLoop()
	return m
}

func (m *Monitor) resetLoop() {
	ticker := time.NewTicker(spamWindow)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			m.counts = make(map[chanKey]int)
			m.mu.Unlock()
		case <-m.done:
			return
		}
	}
}

// Check incrémente le compteur du user dans ce canal et applique le timeout si dépassement.
func (m *Monitor) Check(s *discordgo.Session, guildID, channelID, userID, username string) {
	ck := chanKey{guildID: guildID, channelID: channelID, userID: userID}
	uk := userKey{guildID: guildID, userID: userID}

	m.mu.Lock()
	m.counts[ck]++
	count := m.counts[ck]
	firstOffense := !m.warned[uk]
	if count == spamLimit+1 {
		m.warned[uk] = true
	}
	m.mu.Unlock()

	if count != spamLimit+1 {
		return
	}

	if firstOffense {
		warn := fmt.Sprintf("⚠️ <@%s>, tu envoies trop de messages trop vite. Un timeout de **5 minutes** va être appliqué.", userID)
		if _, err := s.ChannelMessageSend(channelID, warn); err != nil {
			log.Printf("[antispam] Avertissement échoué pour %s: %v", userID, err)
		}
	}

	until := time.Now().Add(timeoutDuration)
	if err := s.GuildMemberTimeout(guildID, userID, &until); err != nil {
		log.Printf("[antispam] Timeout échoué pour %s: %v", userID, err)
	}

	if m.logger != nil {
		m.logger.LogTimeout(s, userID, username, channelID)
	}

	log.Printf("[antispam] Timeout appliqué à %s (%s) dans le canal %s", username, userID, channelID)
}

func (m *Monitor) Shutdown() {
	close(m.done)
}
