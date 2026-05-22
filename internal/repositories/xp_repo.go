package repositories

import (
	"database/sql"

	"discord-bot/internal/database"
)

type UserXP struct {
	UserID   string
	GuildID  string
	TotalXP  int64
	Level    int
	Prestige int
}

type LeaderboardEntry struct {
	UserID  string
	TotalXP int64
	Level   int
	Rank    int64
}

type XPRepo struct {
	db *database.DB
}

func NewXPRepo(db *database.DB) *XPRepo { return &XPRepo{db: db} }

func (r *XPRepo) GetOrCreate(guildID, userID string) (*UserXP, error) {
	_, err := r.db.Exec(`
		INSERT OR IGNORE INTO user_xp (user_id, guild_id, total_xp, level, prestige)
		VALUES (?, ?, 0, 0, 0)
	`, userID, guildID)
	if err != nil {
		return nil, err
	}
	return r.Get(guildID, userID)
}

func (r *XPRepo) Get(guildID, userID string) (*UserXP, error) {
	u := &UserXP{UserID: userID, GuildID: guildID}
	err := r.db.QueryRow(
		`SELECT user_id, guild_id, total_xp, level, prestige FROM user_xp WHERE guild_id=? AND user_id=?`,
		guildID, userID,
	).Scan(&u.UserID, &u.GuildID, &u.TotalXP, &u.Level, &u.Prestige)
	if err == sql.ErrNoRows {
		return u, nil
	}
	return u, err
}

// AddXP ajoute de l'XP et retourne le nouveau total.
func (r *XPRepo) AddXP(guildID, userID string, amount int64) (int64, error) {
	_, err := r.db.Exec(`
		INSERT INTO user_xp (user_id, guild_id, total_xp, level, prestige)
		VALUES (?, ?, ?, 0, 0)
		ON CONFLICT(user_id, guild_id) DO UPDATE SET total_xp = total_xp + excluded.total_xp
	`, userID, guildID, amount)
	if err != nil {
		return 0, err
	}
	var total int64
	err = r.db.QueryRow(`SELECT total_xp FROM user_xp WHERE guild_id=? AND user_id=?`, guildID, userID).Scan(&total)
	return total, err
}

func (r *XPRepo) SetXP(guildID, userID string, totalXP int64, level int) error {
	_, err := r.db.Exec(`
		INSERT INTO user_xp (user_id, guild_id, total_xp, level, prestige)
		VALUES (?, ?, ?, ?, 0)
		ON CONFLICT(user_id, guild_id) DO UPDATE SET total_xp=excluded.total_xp, level=excluded.level
	`, userID, guildID, totalXP, level)
	return err
}

func (r *XPRepo) UpdateLevel(guildID, userID string, level int) error {
	_, err := r.db.Exec(`UPDATE user_xp SET level=? WHERE guild_id=? AND user_id=?`, level, guildID, userID)
	return err
}

func (r *XPRepo) Prestige(guildID, userID string) error {
	_, err := r.db.Exec(
		`UPDATE user_xp SET total_xp=0, level=0, prestige=prestige+1 WHERE guild_id=? AND user_id=?`,
		guildID, userID,
	)
	return err
}

func (r *XPRepo) GetLeaderboard(guildID string, limit, offset int) ([]LeaderboardEntry, error) {
	rows, err := r.db.Query(`
		SELECT user_id, total_xp, level,
		       ROW_NUMBER() OVER (ORDER BY total_xp DESC) AS rank
		FROM user_xp WHERE guild_id=?
		ORDER BY total_xp DESC LIMIT ? OFFSET ?
	`, guildID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LeaderboardEntry
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.TotalXP, &e.Level, &e.Rank); err != nil {
			continue
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *XPRepo) GetRank(guildID, userID string) (rank, total int64, err error) {
	err = r.db.QueryRow(`
		SELECT
			(SELECT COUNT(*)+1 FROM user_xp WHERE guild_id=? AND total_xp >
				COALESCE((SELECT total_xp FROM user_xp WHERE guild_id=? AND user_id=?), 0)),
			(SELECT COUNT(*) FROM user_xp WHERE guild_id=?)
	`, guildID, guildID, userID, guildID).Scan(&rank, &total)
	return
}

func (r *XPRepo) Count(guildID string) (int64, error) {
	var n int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM user_xp WHERE guild_id=?`, guildID).Scan(&n)
	return n, err
}

func (r *XPRepo) Reset(guildID, userID string) error {
	_, err := r.db.Exec(`DELETE FROM user_xp WHERE guild_id=? AND user_id=?`, guildID, userID)
	return err
}
