package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"discord-bot/internal/database"
)

type UserEconomy struct {
	UserID      string
	GuildID     string
	Wallet      int64
	Bank        int64
	DailyStreak int
	LastDaily   *time.Time
	LastWork    *time.Time
}

type EconomyRepo struct {
	db *database.DB
}

func NewEconomyRepo(db *database.DB) *EconomyRepo { return &EconomyRepo{db: db} }

func (r *EconomyRepo) GetOrCreate(guildID, userID string) (*UserEconomy, error) {
	_, err := r.db.Exec(`
		INSERT OR IGNORE INTO economy (user_id, guild_id, wallet, bank, daily_streak)
		VALUES (?, ?, 0, 100, 0)
	`, userID, guildID)
	if err != nil {
		return nil, err
	}
	return r.Get(guildID, userID)
}

func (r *EconomyRepo) Get(guildID, userID string) (*UserEconomy, error) {
	u := &UserEconomy{UserID: userID, GuildID: guildID}
	var ld, lw sql.NullTime
	err := r.db.QueryRow(`
		SELECT user_id, guild_id, wallet, bank, daily_streak, last_daily, last_work
		FROM economy WHERE guild_id=? AND user_id=?
	`, guildID, userID).Scan(&u.UserID, &u.GuildID, &u.Wallet, &u.Bank, &u.DailyStreak, &ld, &lw)
	if err == sql.ErrNoRows {
		return u, nil
	}
	if err != nil {
		return nil, err
	}
	if ld.Valid {
		u.LastDaily = &ld.Time
	}
	if lw.Valid {
		u.LastWork = &lw.Time
	}
	return u, nil
}

func (r *EconomyRepo) AddToWallet(guildID, userID string, amount int64) error {
	_, err := r.db.Exec(`
		INSERT INTO economy (user_id, guild_id, wallet, bank)
		VALUES (?, ?, MAX(0,?), 100)
		ON CONFLICT(user_id, guild_id) DO UPDATE SET wallet = MAX(0, wallet + excluded.wallet)
	`, userID, guildID, amount)
	return err
}

func (r *EconomyRepo) RemoveFromWallet(guildID, userID string, amount int64) error {
	var wallet int64
	if err := r.db.QueryRow(`SELECT COALESCE(wallet,0) FROM economy WHERE guild_id=? AND user_id=?`, guildID, userID).Scan(&wallet); err != nil {
		return fmt.Errorf("compte introuvable")
	}
	if wallet < amount {
		return fmt.Errorf("solde insuffisant (%d 🪙 dans le portefeuille)", wallet)
	}
	_, err := r.db.Exec(`UPDATE economy SET wallet=wallet-? WHERE guild_id=? AND user_id=?`, amount, guildID, userID)
	return err
}

func (r *EconomyRepo) Deposit(guildID, userID string, amount int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var wallet int64
	if err := tx.QueryRow(`SELECT COALESCE(wallet,0) FROM economy WHERE guild_id=? AND user_id=?`, guildID, userID).Scan(&wallet); err != nil {
		return fmt.Errorf("compte introuvable")
	}
	if wallet < amount {
		return fmt.Errorf("portefeuille insuffisant (%d 🪙)", wallet)
	}
	_, err = tx.Exec(`UPDATE economy SET wallet=wallet-?, bank=bank+? WHERE guild_id=? AND user_id=?`, amount, amount, guildID, userID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *EconomyRepo) Withdraw(guildID, userID string, amount int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var bank int64
	if err := tx.QueryRow(`SELECT COALESCE(bank,0) FROM economy WHERE guild_id=? AND user_id=?`, guildID, userID).Scan(&bank); err != nil {
		return fmt.Errorf("compte introuvable")
	}
	if bank < amount {
		return fmt.Errorf("banque insuffisante (%d 🪙)", bank)
	}
	_, err = tx.Exec(`UPDATE economy SET bank=bank-?, wallet=wallet+? WHERE guild_id=? AND user_id=?`, amount, amount, guildID, userID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *EconomyRepo) Transfer(guildID, fromID, toID string, amount int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var wallet int64
	if err := tx.QueryRow(`SELECT COALESCE(wallet,0) FROM economy WHERE guild_id=? AND user_id=?`, guildID, fromID).Scan(&wallet); err != nil || wallet < amount {
		return fmt.Errorf("portefeuille insuffisant (%d 🪙)", wallet)
	}
	if _, err = tx.Exec(`UPDATE economy SET wallet=wallet-? WHERE guild_id=? AND user_id=?`, amount, guildID, fromID); err != nil {
		return err
	}
	if _, err = tx.Exec(`
		INSERT INTO economy (user_id, guild_id, wallet, bank) VALUES (?,?,?,100)
		ON CONFLICT(user_id, guild_id) DO UPDATE SET wallet=wallet+excluded.wallet
	`, toID, guildID, amount); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *EconomyRepo) SetLastDaily(guildID, userID string, streak int, t time.Time) error {
	_, err := r.db.Exec(`UPDATE economy SET daily_streak=?, last_daily=? WHERE guild_id=? AND user_id=?`, streak, t, guildID, userID)
	return err
}

func (r *EconomyRepo) SetLastWork(guildID, userID string, t time.Time) error {
	_, err := r.db.Exec(`UPDATE economy SET last_work=? WHERE guild_id=? AND user_id=?`, t, guildID, userID)
	return err
}

func (r *EconomyRepo) AdminSet(guildID, userID string, wallet, bank int64) error {
	_, err := r.db.Exec(`
		INSERT INTO economy (user_id, guild_id, wallet, bank) VALUES (?,?,?,?)
		ON CONFLICT(user_id, guild_id) DO UPDATE SET wallet=excluded.wallet, bank=excluded.bank
	`, userID, guildID, wallet, bank)
	return err
}

func (r *EconomyRepo) Reset(guildID, userID string) error {
	_, err := r.db.Exec(`
		UPDATE economy SET wallet=0, bank=100, daily_streak=0, last_daily=NULL, last_work=NULL
		WHERE guild_id=? AND user_id=?
	`, guildID, userID)
	return err
}

func (r *EconomyRepo) GetLeaderboard(guildID string, limit, offset int) ([]UserEconomy, error) {
	rows, err := r.db.Query(`
		SELECT user_id, wallet, bank FROM economy WHERE guild_id=?
		ORDER BY (wallet+bank) DESC LIMIT ? OFFSET ?
	`, guildID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []UserEconomy
	for rows.Next() {
		var e UserEconomy
		e.GuildID = guildID
		if err := rows.Scan(&e.UserID, &e.Wallet, &e.Bank); err != nil {
			continue
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
