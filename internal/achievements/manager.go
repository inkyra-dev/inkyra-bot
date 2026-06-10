package achievements

import (
	"log"

	"discord-bot/internal/repositories"
)

// Achievement describes a single achievement in the catalogue.
type Achievement struct {
	Key         string
	Name        string
	Description string
	Emoji       string
}

// All is the static catalogue. Order determines display order in /achievements.
var All = []Achievement{
	// Messages & XP
	{Key: "first_message", Name: "Premier pas", Description: "Envoie ton premier message", Emoji: "💬"},
	{Key: "msg_100", Name: "Bavard", Description: "Envoie 100 messages", Emoji: "🗣️"},
	{Key: "msg_500", Name: "Intarissable", Description: "Envoie 500 messages", Emoji: "📢"},
	{Key: "msg_1000", Name: "Légende du chat", Description: "Envoie 1000 messages", Emoji: "🏆"},
	{Key: "level_5", Name: "Apprenti", Description: "Atteins le niveau 5", Emoji: "⭐"},
	{Key: "level_10", Name: "Confirmé", Description: "Atteins le niveau 10", Emoji: "🌟"},
	{Key: "level_25", Name: "Expérimenté", Description: "Atteins le niveau 25", Emoji: "💫"},
	{Key: "level_50", Name: "Vétéran", Description: "Atteins le niveau 50", Emoji: "🏅"},
	{Key: "level_100", Name: "Maître", Description: "Atteins le niveau 100", Emoji: "👑"},
	{Key: "prestige", Name: "Prestige", Description: "Effectue ton premier prestige", Emoji: "🌠"},
	// Économie
	{Key: "first_daily", Name: "Ponctuel", Description: "Récupère ta première récompense quotidienne", Emoji: "📅"},
	{Key: "daily_streak_7", Name: "Régulier", Description: "Maintiens une série de 7 jours consécutifs", Emoji: "🔥"},
	{Key: "first_purchase", Name: "Premier achat", Description: "Achète un item dans la boutique", Emoji: "🛒"},
	{Key: "millionaire", Name: "Millionnaire", Description: "Accumule 1 000 000 🪙 (portefeuille + banque)", Emoji: "💰"},
	// Blackjack
	{Key: "first_blackjack", Name: "Joueur", Description: "Lance une première partie de blackjack", Emoji: "🃏"},
	{Key: "blackjack_win", Name: "Chanceux", Description: "Gagne une partie de blackjack", Emoji: "🎉"},
	{Key: "blackjack_natural", Name: "Naturel", Description: "Fais un blackjack naturel (21 dès le départ)", Emoji: "🎰"},
	{Key: "bj_win_10", Name: "Sharp", Description: "Gagne 10 parties de blackjack", Emoji: "🎯"},
	{Key: "bj_win_50", Name: "Card Shark", Description: "Gagne 50 parties de blackjack", Emoji: "🦈"},
	// Jeux
	{Key: "slots_jackpot", Name: "Lucky", Description: "Décroches le jackpot maximum aux slots (3×7️⃣)", Emoji: "🍀"},
	{Key: "dice_double6", Name: "Snake Eyes", Description: "Obtiens un 6 aux dés", Emoji: "🎲"},
	// Social
	{Key: "welcome", Name: "Bienvenue", Description: "Rejoins le serveur", Emoji: "👋"},
}

// byKey is a lookup map built once at init time.
var byKey map[string]Achievement

func init() {
	byKey = make(map[string]Achievement, len(All))
	for _, a := range All {
		byKey[a.Key] = a
	}
}

// Lookup returns an achievement by key. The second return value is false if not found.
func Lookup(key string) (Achievement, bool) {
	a, ok := byKey[key]
	return a, ok
}

// ── Manager ───────────────────────────────────────────────────────────────────

type Manager struct {
	repo *repositories.AchievementRepo
}

func NewManager(repo *repositories.AchievementRepo) *Manager {
	return &Manager{repo: repo}
}

// Check tries to unlock an achievement. Errors are logged and never propagated —
// achievement failures must never block gameplay.
func (m *Manager) Check(guildID, userID, key string) {
	if _, err := m.repo.Unlock(guildID, userID, key); err != nil {
		log.Printf("[achievements] unlock %q for %s/%s: %v", key, guildID, userID, err)
	}
}

// AchievementStatus pairs a catalogue entry with whether it has been unlocked.
type AchievementStatus struct {
	Achievement
	Unlocked bool
}

// GetStatus returns every achievement in catalogue order, with its unlock state for
// the given user.
func (m *Manager) GetStatus(guildID, userID string) ([]AchievementStatus, error) {
	keys, err := m.repo.GetByUser(guildID, userID)
	if err != nil {
		return nil, err
	}
	unlocked := make(map[string]bool, len(keys))
	for _, k := range keys {
		unlocked[k] = true
	}
	out := make([]AchievementStatus, len(All))
	for i, a := range All {
		out[i] = AchievementStatus{Achievement: a, Unlocked: unlocked[a.Key]}
	}
	return out, nil
}
