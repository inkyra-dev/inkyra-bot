package games

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ── Cartes ────────────────────────────────────────────────────────────────────

type Card struct {
	Suit  string
	Value string
}

func (c Card) String() string { return c.Value + c.Suit }

func cardPoints(c Card) int {
	switch c.Value {
	case "A":
		return 11
	case "J", "Q", "K":
		return 10
	default:
		v, _ := strconv.Atoi(c.Value)
		return v
	}
}

func handValue(hand []Card) int {
	total, aces := 0, 0
	for _, c := range hand {
		v := cardPoints(c)
		if v == 11 {
			aces++
		}
		total += v
	}
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}
	return total
}

func handStr(hand []Card, hideSecond bool) string {
	parts := make([]string, len(hand))
	for i, c := range hand {
		if hideSecond && i == 1 {
			parts[i] = "🂠"
		} else {
			parts[i] = "`" + c.String() + "`"
		}
	}
	return strings.Join(parts, " ")
}

func newDeck() []Card {
	suits := []string{"♠", "♥", "♦", "♣"}
	values := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	deck := make([]Card, 0, 52)
	for _, s := range suits {
		for _, v := range values {
			deck = append(deck, Card{Suit: s, Value: v})
		}
	}
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

// ── Session ───────────────────────────────────────────────────────────────────

type BJOutcome int

const (
	BJOngoing BJOutcome = iota
	BJPlayerWin
	BJDealerWin
	BJBlackjack // blackjack naturel du joueur
	BJPush      // égalité
	BJBust      // joueur dépasse 21
)

type BlackjackSession struct {
	UserID     string
	GuildID    string
	ChannelID  string
	MessageID  string
	Bet        int64
	Deck       []Card
	PlayerHand []Card
	DealerHand []Card
	CreatedAt  time.Time
	Done       bool
}

func newBJSession(guildID, userID, channelID string, bet int64) *BlackjackSession {
	deck := newDeck()
	s := &BlackjackSession{
		UserID:     userID,
		GuildID:    guildID,
		ChannelID:  channelID,
		Bet:        bet,
		Deck:       deck[4:],
		PlayerHand: []Card{deck[0], deck[2]},
		DealerHand: []Card{deck[1], deck[3]},
		CreatedAt:  time.Now(),
	}
	return s
}

func (s *BlackjackSession) draw() Card {
	c := s.Deck[0]
	s.Deck = s.Deck[1:]
	return c
}

func (s *BlackjackSession) PlayerScore() int  { return handValue(s.PlayerHand) }
func (s *BlackjackSession) DealerScore() int  { return handValue(s.DealerHand) }

func (s *BlackjackSession) PlayerHandStr(reveal bool) string {
	return handStr(s.PlayerHand, false) + fmt.Sprintf(" **(score: %d)**", s.PlayerScore())
}

func (s *BlackjackSession) DealerHandStr(reveal bool) string {
	return handStr(s.DealerHand, !reveal) + func() string {
		if reveal {
			return fmt.Sprintf(" **(score: %d)**", s.DealerScore())
		}
		return fmt.Sprintf(" **(score: %d + ?)**", cardPoints(s.DealerHand[0]))
	}()
}

// initialOutcome vérifie si le joueur a un blackjack naturel au départ.
func (s *BlackjackSession) initialOutcome() BJOutcome {
	if s.PlayerScore() == 21 {
		return BJBlackjack
	}
	return BJOngoing
}

// Hit ajoute une carte au joueur.
func (s *BlackjackSession) Hit() BJOutcome {
	s.PlayerHand = append(s.PlayerHand, s.draw())
	if s.PlayerScore() > 21 {
		s.Done = true
		return BJBust
	}
	if s.PlayerScore() == 21 {
		return s.Stand()
	}
	return BJOngoing
}

// Stand joue pour le dealer (règles : tire jusqu'à 17).
func (s *BlackjackSession) Stand() BJOutcome {
	s.Done = true
	for s.DealerScore() < 17 {
		s.DealerHand = append(s.DealerHand, s.draw())
	}
	ps, ds := s.PlayerScore(), s.DealerScore()
	switch {
	case ds > 21 || ps > ds:
		return BJPlayerWin
	case ps < ds:
		return BJDealerWin
	default:
		return BJPush
	}
}

// Payout retourne le gain net pour un résultat donné.
func (s *BlackjackSession) Payout(outcome BJOutcome) int64 {
	switch outcome {
	case BJBlackjack:
		return int64(float64(s.Bet) * 1.5) // 3:2
	case BJPlayerWin:
		return s.Bet
	case BJPush:
		return 0 // remboursement (géré par l'appelant)
	default:
		return -s.Bet
	}
}

// ── Manager ───────────────────────────────────────────────────────────────────

type BJManager struct {
	mu       sync.RWMutex
	sessions map[string]*BlackjackSession // key: guildID+":"+userID
}

func NewBJManager() *BJManager {
	bj := &BJManager{sessions: make(map[string]*BlackjackSession)}
	go bj.cleanup()
	return bj
}

func (m *BJManager) key(guildID, userID string) string { return guildID + ":" + userID }

func (m *BJManager) Start(guildID, userID, channelID string, bet int64) (*BlackjackSession, BJOutcome) {
	key := m.key(guildID, userID)
	s := newBJSession(guildID, userID, channelID, bet)

	m.mu.Lock()
	m.sessions[key] = s
	m.mu.Unlock()

	return s, s.initialOutcome()
}

func (m *BJManager) Get(guildID, userID string) (*BlackjackSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[m.key(guildID, userID)]
	return s, ok
}

func (m *BJManager) Delete(guildID, userID string) {
	m.mu.Lock()
	delete(m.sessions, m.key(guildID, userID))
	m.mu.Unlock()
}

func (m *BJManager) HasActive(guildID, userID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[m.key(guildID, userID)]
	return ok && !s.Done
}

func (m *BJManager) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		limit := time.Now().Add(-10 * time.Minute)
		m.mu.Lock()
		for k, s := range m.sessions {
			if s.CreatedAt.Before(limit) {
				delete(m.sessions, k)
			}
		}
		m.mu.Unlock()
	}
}
