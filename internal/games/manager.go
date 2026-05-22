package games

import "discord-bot/internal/repositories"

// Manager centralise tous les jeux et l'accès à l'économie.
type Manager struct {
	BJ       *BJManager
	econRepo *repositories.EconomyRepo
}

func NewManager(econRepo *repositories.EconomyRepo) *Manager {
	return &Manager{
		BJ:       NewBJManager(),
		econRepo: econRepo,
	}
}

// Debit retire `amount` du portefeuille de l'utilisateur (mise de jeu).
func (m *Manager) Debit(guildID, userID string, amount int64) error {
	return m.econRepo.RemoveFromWallet(guildID, userID, amount)
}

// Credit ajoute `amount` au portefeuille de l'utilisateur (gain).
func (m *Manager) Credit(guildID, userID string, amount int64) error {
	return m.econRepo.AddToWallet(guildID, userID, amount)
}

// Refund restitue exactement la mise (push/égalité).
func (m *Manager) Refund(guildID, userID string, amount int64) error {
	return m.econRepo.AddToWallet(guildID, userID, amount)
}
