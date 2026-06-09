package repositories

import "discord-bot/internal/database"

type AchievementRepo struct {
	db *database.DB
}

func NewAchievementRepo(db *database.DB) *AchievementRepo {
	return &AchievementRepo{db: db}
}

// Unlock inserts the achievement if not already present. Returns true if newly unlocked.
func (r *AchievementRepo) Unlock(guildID, userID, achievement string) (bool, error) {
	res, err := r.db.Exec(`
		INSERT OR IGNORE INTO achievements (user_id, guild_id, achievement)
		VALUES (?, ?, ?)
	`, userID, guildID, achievement)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// GetByUser returns the keys of all achievements unlocked by a user, ordered by earn date.
func (r *AchievementRepo) GetByUser(guildID, userID string) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT achievement FROM achievements WHERE guild_id=? AND user_id=? ORDER BY earned_at
	`, guildID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			continue
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// IsUnlocked checks whether a specific achievement is already unlocked.
func (r *AchievementRepo) IsUnlocked(guildID, userID, achievement string) (bool, error) {
	var n int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM achievements WHERE guild_id=? AND user_id=? AND achievement=?
	`, guildID, userID, achievement).Scan(&n)
	return n > 0, err
}
