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
	`)
	return err
}
