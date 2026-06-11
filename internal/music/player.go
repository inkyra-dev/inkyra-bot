package music

import (
	"log/slog"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// Player gère la lecture audio pour un seul serveur.
type Player struct {
	mu      sync.Mutex
	vc      *discordgo.VoiceConnection
	queue   *Queue
	stop    chan struct{}
	skip    chan struct{}
	paused  bool
	playing bool
	current *Track
}

func newPlayer() *Player {
	return &Player{
		queue: &Queue{},
		stop:  make(chan struct{}, 1),
		skip:  make(chan struct{}, 1),
	}
}

func (p *Player) Connect(s *discordgo.Session, guildID, channelID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}
	p.vc = vc
	return nil
}

func (p *Player) Disconnect() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.vc != nil {
		p.vc.Disconnect()
		p.vc = nil
	}
}

func (p *Player) Enqueue(t Track) {
	p.queue.Add(t)
}

// Start lance la boucle de lecture si elle n'est pas déjà active.
func (p *Player) Start(s *discordgo.Session, guildID string) {
	p.mu.Lock()
	if p.playing {
		p.mu.Unlock()
		return
	}
	p.playing = true
	p.mu.Unlock()

	go p.loop(s, guildID)
}

func (p *Player) loop(s *discordgo.Session, guildID string) {
	defer func() {
		p.mu.Lock()
		p.playing = false
		p.current = nil
		p.mu.Unlock()
	}()

	for {
		track, ok := p.queue.Pop()
		if !ok {
			// Queue vide, on déconnecte automatiquement
			p.Disconnect()
			return
		}

		p.mu.Lock()
		p.current = &track
		p.mu.Unlock()

		slog.Info("lecture", "component", "music", "title", track.Title)

		done := make(chan bool)
		go playAudioFile(p.vc, track.URL, done)

		select {
		case <-done:
			// Lecture terminée normalement, on passe à la suivante
		case <-p.skip:
			// Skip demandé
			done <- true
		case <-p.stop:
			// Stop total
			done <- true
			p.queue.Clear()
			p.Disconnect()
			return
		}
	}
}

func (p *Player) Skip() {
	select {
	case p.skip <- struct{}{}:
	default:
	}
}

func (p *Player) Stop() {
	p.queue.Clear()
	select {
	case p.stop <- struct{}{}:
	default:
		// Pas de lecture en cours, juste déconnecter
		p.Disconnect()
	}
}

func (p *Player) IsPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playing
}

func (p *Player) Current() *Track {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current
}

func (p *Player) VoiceChannelID() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.vc == nil {
		return ""
	}
	return p.vc.ChannelID
}

func (p *Player) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.vc != nil {
		p.vc.Speaking(false)
		p.paused = true
	}
}

func (p *Player) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.vc != nil {
		p.vc.Speaking(true)
		p.paused = false
	}
}

func (p *Player) Queue() *Queue {
	return p.queue
}
