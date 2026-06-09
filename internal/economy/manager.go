package economy

import (
	"fmt"
	"math/rand"
	"time"

	"discord-bot/internal/achievements"
	"discord-bot/internal/repositories"
)

var jobs = []struct {
	name   string
	action string
}{
	{"développeur", "corrigé des bugs"},
	{"livreur", "livré des colis"},
	{"plombier", "réparé des tuyaux"},
	{"jardinier", "taillé des haies"},
	{"cuisinier", "préparé des repas"},
	{"chauffeur", "conduit un client"},
	{"technicien", "réparé des appareils"},
	{"médecin", "soigné des patients"},
	{"streamer", "collecté des dons"},
	{"trader", "fait des trades"},
}

type Manager struct {
	econRepo *repositories.EconomyRepo
	itemRepo *repositories.ItemsRepo
	achMgr   *achievements.Manager
}

func NewManager(econRepo *repositories.EconomyRepo, itemRepo *repositories.ItemsRepo, achMgr *achievements.Manager) *Manager {
	return &Manager{econRepo: econRepo, itemRepo: itemRepo, achMgr: achMgr}
}

func (m *Manager) checkMillionaire(guildID, userID string) {
	u, err := m.econRepo.Get(guildID, userID)
	if err != nil {
		return
	}
	if u.Wallet+u.Bank >= 1_000_000 {
		m.achMgr.Check(guildID, userID, "millionaire")
	}
}

func (m *Manager) Balance(guildID, userID string) (*repositories.UserEconomy, error) {
	return m.econRepo.GetOrCreate(guildID, userID)
}

type DailyResult struct {
	Reward int64
	Streak int
}

func (m *Manager) Daily(guildID, userID string) (*DailyResult, error) {
	u, err := m.econRepo.GetOrCreate(guildID, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if u.LastDaily != nil && now.Sub(*u.LastDaily) < 24*time.Hour {
		remaining := 24*time.Hour - now.Sub(*u.LastDaily)
		return nil, fmt.Errorf("revenez dans **%s**", formatDuration(remaining))
	}

	streak := 1
	if u.LastDaily != nil && now.Sub(*u.LastDaily) < 48*time.Hour {
		streak = u.DailyStreak + 1
	}

	reward := int64(200) + int64(streak)*10
	if reward > 600 {
		reward = 600
	}

	if err := m.econRepo.AddToWallet(guildID, userID, reward); err != nil {
		return nil, err
	}
	if err := m.econRepo.SetLastDaily(guildID, userID, streak, now); err != nil {
		return nil, err
	}
	m.achMgr.Check(guildID, userID, "first_daily")
	if streak >= 7 {
		m.achMgr.Check(guildID, userID, "daily_streak_7")
	}
	m.checkMillionaire(guildID, userID)
	return &DailyResult{Reward: reward, Streak: streak}, nil
}

type WorkResult struct {
	Reward int64
	Job    string
	Action string
}

func (m *Manager) Work(guildID, userID string) (*WorkResult, error) {
	u, err := m.econRepo.GetOrCreate(guildID, userID)
	if err != nil {
		return nil, err
	}

	if u.LastWork != nil && time.Since(*u.LastWork) < 4*time.Hour {
		remaining := 4*time.Hour - time.Since(*u.LastWork)
		return nil, fmt.Errorf("vous avez déjà travaillé, revenez dans **%s**", formatDuration(remaining))
	}

	reward := int64(50 + rand.Intn(151)) // 50–200
	job := jobs[rand.Intn(len(jobs))]

	if err := m.econRepo.AddToWallet(guildID, userID, reward); err != nil {
		return nil, err
	}
	if err := m.econRepo.SetLastWork(guildID, userID, time.Now()); err != nil {
		return nil, err
	}
	m.checkMillionaire(guildID, userID)
	return &WorkResult{Reward: reward, Job: job.name, Action: job.action}, nil
}

func (m *Manager) Deposit(guildID, userID string, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("montant invalide")
	}
	return m.econRepo.Deposit(guildID, userID, amount)
}

func (m *Manager) DepositAll(guildID, userID string) (int64, error) {
	u, err := m.econRepo.Get(guildID, userID)
	if err != nil || u.Wallet == 0 {
		return 0, fmt.Errorf("portefeuille vide")
	}
	return u.Wallet, m.econRepo.Deposit(guildID, userID, u.Wallet)
}

func (m *Manager) Withdraw(guildID, userID string, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("montant invalide")
	}
	return m.econRepo.Withdraw(guildID, userID, amount)
}

func (m *Manager) WithdrawAll(guildID, userID string) (int64, error) {
	u, err := m.econRepo.Get(guildID, userID)
	if err != nil || u.Bank == 0 {
		return 0, fmt.Errorf("banque vide")
	}
	return u.Bank, m.econRepo.Withdraw(guildID, userID, u.Bank)
}

func (m *Manager) Pay(guildID, fromID, toID string, amount int64) error {
	if fromID == toID {
		return fmt.Errorf("tu ne peux pas te payer toi-même")
	}
	if amount <= 0 {
		return fmt.Errorf("montant invalide")
	}
	return m.econRepo.Transfer(guildID, fromID, toID, amount)
}

func (m *Manager) Shop(guildID string) ([]repositories.Item, error) {
	return m.itemRepo.GetAll(guildID)
}

func (m *Manager) Buy(guildID, userID string, itemID int64) (*repositories.Item, error) {
	item, err := m.itemRepo.GetByID(itemID)
	if err != nil {
		return nil, err
	}
	if err := m.itemRepo.Buy(guildID, userID, itemID); err != nil {
		return nil, err
	}
	m.achMgr.Check(guildID, userID, "first_purchase")
	return item, nil
}

func (m *Manager) Sell(guildID, userID string, itemID int64) (*repositories.Item, error) {
	item, err := m.itemRepo.GetByID(itemID)
	if err != nil {
		return nil, err
	}
	return item, m.itemRepo.Sell(guildID, userID, itemID)
}

func (m *Manager) Inventory(guildID, userID string) ([]repositories.InventoryItem, error) {
	return m.itemRepo.GetInventory(guildID, userID)
}

func (m *Manager) AdminGive(guildID, userID string, amount int64) error {
	return m.econRepo.AddToWallet(guildID, userID, amount)
}

func (m *Manager) AdminRemove(guildID, userID string, amount int64) error {
	u, err := m.econRepo.Get(guildID, userID)
	if err != nil {
		return err
	}
	newWallet := u.Wallet - amount
	if newWallet < 0 {
		newWallet = 0
	}
	return m.econRepo.AdminSet(guildID, userID, newWallet, u.Bank)
}

func (m *Manager) AdminReset(guildID, userID string) error {
	return m.econRepo.Reset(guildID, userID)
}

func (m *Manager) GetLeaderboard(guildID string, limit, offset int) ([]repositories.UserEconomy, error) {
	return m.econRepo.GetLeaderboard(guildID, limit, offset)
}

func (m *Manager) Count(guildID string) (int64, error) {
	return m.econRepo.Count(guildID)
}

func (m *Manager) AddToWallet(guildID, userID string, amount int64) error {
	return m.econRepo.AddToWallet(guildID, userID, amount)
}

func (m *Manager) RemoveFromWallet(guildID, userID string, amount int64) error {
	return m.econRepo.RemoveFromWallet(guildID, userID, amount)
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	mn := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %02dm %02ds", h, mn, s)
	}
	if mn > 0 {
		return fmt.Sprintf("%dm %02ds", mn, s)
	}
	return fmt.Sprintf("%ds", s)
}
