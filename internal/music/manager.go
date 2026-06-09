package music

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

// Manager gère un Player par serveur Discord.
type Manager struct {
	mu      sync.Mutex
	players map[string]*Player // clé : guildID
}

func NewManager() *Manager {
	return &Manager{players: make(map[string]*Player)}
}

func (m *Manager) getOrCreate(guildID string) *Player {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.players[guildID]; ok {
		return p
	}
	p := newPlayer()
	m.players[guildID] = p
	return p
}

func (m *Manager) Get(guildID string) (*Player, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.players[guildID]
	return p, ok
}

func (m *Manager) Play(s *discordgo.Session, guildID, voiceChannelID string, track Track) error {
	p := m.getOrCreate(guildID)

	// Rejoindre le vocal si pas encore connecté
	if p.VoiceChannelID() == "" {
		if err := p.Connect(s, guildID, voiceChannelID); err != nil {
			return err
		}
	}

	p.Enqueue(track)
	p.Start(s, guildID)
	return nil
}

func (m *Manager) Shutdown() {
	m.mu.Lock()
	players := make([]*Player, 0, len(m.players))
	for _, p := range m.players {
		players = append(players, p)
	}
	m.mu.Unlock()
	for _, p := range players {
		p.Stop()
	}
}

// HandleVoiceStateUpdate détecte quand le bot est seul dans un canal et se déconnecte.
func (m *Manager) HandleVoiceStateUpdate(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	if vs.GuildID == "" {
		return
	}

	p, ok := m.Get(vs.GuildID)
	if !ok || p.VoiceChannelID() == "" {
		return
	}

	guild, err := s.State.Guild(vs.GuildID)
	if err != nil {
		return
	}

	botChannelID := p.VoiceChannelID()
	humanCount := 0
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == botChannelID && vs.UserID != s.State.User.ID {
			humanCount++
		}
	}

	if humanCount == 0 {
		p.Stop()
	}
}
