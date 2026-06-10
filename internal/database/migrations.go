package database

func (db *DB) migrate() error {
	_, err := db.Exec(`
		-- Tickets
		CREATE TABLE IF NOT EXISTS tickets (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			channel_id TEXT    NOT NULL UNIQUE,
			guild_id   TEXT    NOT NULL,
			user_id    TEXT    NOT NULL,
			status     TEXT    NOT NULL DEFAULT 'open',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			closed_at  DATETIME
		);

		CREATE TABLE IF NOT EXISTS guild_settings (
			guild_id           TEXT PRIMARY KEY,
			ticket_category_id TEXT,
			log_channel_id     TEXT,
			staff_role_id      TEXT
		);

		-- XP
		CREATE TABLE IF NOT EXISTS user_xp (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id    TEXT    NOT NULL,
			guild_id   TEXT    NOT NULL,
			total_xp   INTEGER NOT NULL DEFAULT 0,
			level      INTEGER NOT NULL DEFAULT 0,
			prestige   INTEGER NOT NULL DEFAULT 0,
			UNIQUE(user_id, guild_id)
		);

		-- Economy
		CREATE TABLE IF NOT EXISTS economy (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id      TEXT    NOT NULL,
			guild_id     TEXT    NOT NULL,
			wallet       INTEGER NOT NULL DEFAULT 0,
			bank         INTEGER NOT NULL DEFAULT 100,
			daily_streak INTEGER NOT NULL DEFAULT 0,
			last_daily   DATETIME,
			last_work    DATETIME,
			UNIQUE(user_id, guild_id)
		);

		-- Shop items catalog
		CREATE TABLE IF NOT EXISTS items (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id    TEXT    NOT NULL,
			name        TEXT    NOT NULL,
			description TEXT    NOT NULL DEFAULT '',
			price       INTEGER NOT NULL,
			sell_price  INTEGER NOT NULL DEFAULT 0,
			emoji       TEXT    NOT NULL DEFAULT '📦',
			stock       INTEGER NOT NULL DEFAULT -1
		);

		-- User inventories
		CREATE TABLE IF NOT EXISTS inventory (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id     TEXT    NOT NULL,
			guild_id    TEXT    NOT NULL,
			item_id     INTEGER NOT NULL REFERENCES items(id),
			quantity    INTEGER NOT NULL DEFAULT 1,
			acquired_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, guild_id, item_id)
		);

		-- Achievements
		CREATE TABLE IF NOT EXISTS achievements (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id     TEXT    NOT NULL,
			guild_id    TEXT    NOT NULL,
			achievement TEXT    NOT NULL,
			earned_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, guild_id, achievement)
		);

		-- Blackjack active sessions (purgées au démarrage pour nettoyer les orphelines)
		CREATE TABLE IF NOT EXISTS blackjack_sessions (
			user_id    TEXT    NOT NULL,
			guild_id   TEXT    NOT NULL,
			created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
			PRIMARY KEY (user_id, guild_id)
		);
	`)
	if err != nil {
		return err
	}

	// Purge orphaned blackjack sessions left over from a previous crash or restart.
	if _, err := db.Exec(`DELETE FROM blackjack_sessions`); err != nil {
		return err
	}

	// Add last_xp_at to user_xp for persistent XP cooldowns (idempotent).
	var colExists int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('user_xp') WHERE name='last_xp_at'`).Scan(&colExists); err != nil {
		return err
	}
	if colExists == 0 {
		if _, err := db.Exec(`ALTER TABLE user_xp ADD COLUMN last_xp_at INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
	}

	// Add moderation/auto-roles + config columns to guild_settings (idempotent).
	for _, col := range []struct{ name, typ string }{
		{"auto_role_id", "TEXT"},
		{"roles_channel_id", "TEXT"},
		{"roles_message_id", "TEXT"},
		{"levelup_channel_id", "TEXT"},
		{"daily_cooldown_hours", "INTEGER"},
		{"work_cooldown_hours", "INTEGER"},
		{"max_bet", "INTEGER"},
	} {
		var exists int
		if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('guild_settings') WHERE name=?`, col.name).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			if _, err := db.Exec(`ALTER TABLE guild_settings ADD COLUMN ` + col.name + ` ` + col.typ); err != nil {
				return err
			}
		}
	}

	// Role-selection buttons (idempotent).
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS role_buttons (
			id       INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			role_id  TEXT NOT NULL,
			label    TEXT NOT NULL,
			emoji    TEXT NOT NULL DEFAULT '',
			UNIQUE(guild_id, role_id)
		)
	`); err != nil {
		return err
	}

	// Per-user stats counters for achievement thresholds (idempotent).
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_stats (
			guild_id      TEXT    NOT NULL,
			user_id       TEXT    NOT NULL,
			message_count INTEGER NOT NULL DEFAULT 0,
			bj_wins       INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (guild_id, user_id)
		)
	`)
	return err
}
