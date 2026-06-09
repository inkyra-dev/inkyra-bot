package repositories

import "discord-bot/internal/database"

type BJRepo struct {
	db *database.DB
}

func NewBJRepo(db *database.DB) *BJRepo { return &BJRepo{db: db} }

func (r *BJRepo) Open(guildID, userID string) error {
	_, err := r.db.Exec(`
		INSERT OR REPLACE INTO blackjack_sessions (user_id, guild_id, created_at)
		VALUES (?, ?, strftime('%s', 'now'))
	`, userID, guildID)
	return err
}

func (r *BJRepo) Close(guildID, userID string) error {
	_, err := r.db.Exec(
		`DELETE FROM blackjack_sessions WHERE guild_id=? AND user_id=?`,
		guildID, userID,
	)
	return err
}
