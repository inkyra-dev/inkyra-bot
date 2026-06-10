package repositories

import "discord-bot/internal/database"

type StatsRepo struct {
	db *database.DB
}

func NewStatsRepo(db *database.DB) *StatsRepo { return &StatsRepo{db: db} }

// IncrementMessages increments message_count and returns the new total.
func (r *StatsRepo) IncrementMessages(guildID, userID string) (int64, error) {
	_, err := r.db.Exec(`
		INSERT INTO user_stats (guild_id, user_id, message_count)
		VALUES (?, ?, 1)
		ON CONFLICT(guild_id, user_id) DO UPDATE SET message_count = message_count + 1
	`, guildID, userID)
	if err != nil {
		return 0, err
	}
	var n int64
	err = r.db.QueryRow(
		`SELECT message_count FROM user_stats WHERE guild_id=? AND user_id=?`, guildID, userID,
	).Scan(&n)
	return n, err
}

// IncrementBJWins increments bj_wins and returns the new total.
func (r *StatsRepo) IncrementBJWins(guildID, userID string) (int64, error) {
	_, err := r.db.Exec(`
		INSERT INTO user_stats (guild_id, user_id, bj_wins)
		VALUES (?, ?, 1)
		ON CONFLICT(guild_id, user_id) DO UPDATE SET bj_wins = bj_wins + 1
	`, guildID, userID)
	if err != nil {
		return 0, err
	}
	var n int64
	err = r.db.QueryRow(
		`SELECT bj_wins FROM user_stats WHERE guild_id=? AND user_id=?`, guildID, userID,
	).Scan(&n)
	return n, err
}
