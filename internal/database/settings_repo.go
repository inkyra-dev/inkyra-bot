package database

import (
	"database/sql"
)

// GuildConfig regroupe tous les réglages guild_settings en une seule requête.
type GuildConfig struct {
	GuildID            string
	TicketCategoryID   string
	LogChannelID       string
	StaffRoleID        string
	AutoRoleID         string
	RolesChannelID     string
	LevelUpChannelID   string
	DailyCooldownHours int   // 0 = valeur par défaut (24)
	WorkCooldownHours  int   // 0 = valeur par défaut (4)
	MaxBet             int64 // 0 = illimité
}

type RoleButton struct {
	ID      int64
	GuildID string
	RoleID  string
	Label   string
	Emoji   string
}

// ── guild_settings ────────────────────────────────────────────────────────────

func (db *DB) SetAutoRoleID(guildID, roleID string) error {
	_, err := db.Exec(`
		INSERT INTO guild_settings (guild_id, auto_role_id)
		VALUES (?, ?)
		ON CONFLICT(guild_id) DO UPDATE SET auto_role_id = excluded.auto_role_id
	`, guildID, roleID)
	return err
}

func (db *DB) GetAutoRoleID(guildID string) (string, error) {
	var v sql.NullString
	err := db.QueryRow(`SELECT auto_role_id FROM guild_settings WHERE guild_id=?`, guildID).Scan(&v)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return v.String, err
}

func (db *DB) SetRolesChannelAndMessage(guildID, channelID, messageID string) error {
	_, err := db.Exec(`
		INSERT INTO guild_settings (guild_id, roles_channel_id, roles_message_id)
		VALUES (?, ?, ?)
		ON CONFLICT(guild_id) DO UPDATE
			SET roles_channel_id = excluded.roles_channel_id,
			    roles_message_id  = excluded.roles_message_id
	`, guildID, channelID, messageID)
	return err
}

func (db *DB) GetRolesMessageInfo(guildID string) (channelID, messageID string, err error) {
	var ch, msg sql.NullString
	err = db.QueryRow(
		`SELECT roles_channel_id, roles_message_id FROM guild_settings WHERE guild_id=?`, guildID,
	).Scan(&ch, &msg)
	if err == sql.ErrNoRows {
		return "", "", nil
	}
	return ch.String, msg.String, err
}

// ── level-up channel ──────────────────────────────────────────────────────────

func (db *DB) SetLevelUpChannelID(guildID, channelID string) error {
	_, err := db.Exec(`
		INSERT INTO guild_settings (guild_id, levelup_channel_id)
		VALUES (?, ?)
		ON CONFLICT(guild_id) DO UPDATE SET levelup_channel_id = excluded.levelup_channel_id
	`, guildID, channelID)
	return err
}

func (db *DB) GetLevelUpChannelID(guildID string) (string, error) {
	var v sql.NullString
	err := db.QueryRow(`SELECT levelup_channel_id FROM guild_settings WHERE guild_id=?`, guildID).Scan(&v)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return v.String, err
}

// ── cooldowns configurables ───────────────────────────────────────────────────

func (db *DB) SetDailyCooldownHours(guildID string, hours int) error {
	_, err := db.Exec(`
		INSERT INTO guild_settings (guild_id, daily_cooldown_hours)
		VALUES (?, ?)
		ON CONFLICT(guild_id) DO UPDATE SET daily_cooldown_hours = excluded.daily_cooldown_hours
	`, guildID, hours)
	return err
}

func (db *DB) GetDailyCooldownHours(guildID string) (int, error) {
	var v sql.NullInt64
	err := db.QueryRow(`SELECT daily_cooldown_hours FROM guild_settings WHERE guild_id=?`, guildID).Scan(&v)
	if err == sql.ErrNoRows || !v.Valid || v.Int64 == 0 {
		return 0, nil
	}
	return int(v.Int64), err
}

func (db *DB) SetWorkCooldownHours(guildID string, hours int) error {
	_, err := db.Exec(`
		INSERT INTO guild_settings (guild_id, work_cooldown_hours)
		VALUES (?, ?)
		ON CONFLICT(guild_id) DO UPDATE SET work_cooldown_hours = excluded.work_cooldown_hours
	`, guildID, hours)
	return err
}

func (db *DB) GetWorkCooldownHours(guildID string) (int, error) {
	var v sql.NullInt64
	err := db.QueryRow(`SELECT work_cooldown_hours FROM guild_settings WHERE guild_id=?`, guildID).Scan(&v)
	if err == sql.ErrNoRows || !v.Valid || v.Int64 == 0 {
		return 0, nil
	}
	return int(v.Int64), err
}

// ── mise maximale ─────────────────────────────────────────────────────────────

func (db *DB) SetMaxBet(guildID string, amount int64) error {
	_, err := db.Exec(`
		INSERT INTO guild_settings (guild_id, max_bet)
		VALUES (?, ?)
		ON CONFLICT(guild_id) DO UPDATE SET max_bet = excluded.max_bet
	`, guildID, amount)
	return err
}

// GetMaxBet retourne la mise max (0 = illimité).
func (db *DB) GetMaxBet(guildID string) (int64, error) {
	var v sql.NullInt64
	err := db.QueryRow(`SELECT max_bet FROM guild_settings WHERE guild_id=?`, guildID).Scan(&v)
	if err == sql.ErrNoRows || !v.Valid {
		return 0, nil
	}
	return v.Int64, err
}

// ── vue complète pour /config ─────────────────────────────────────────────────

func (db *DB) GetGuildConfig(guildID string) (*GuildConfig, error) {
	cfg := &GuildConfig{GuildID: guildID}
	var (
		ticketCat, logChan, staffRole sql.NullString
		autoRole, rolesChan, lvlupChan sql.NullString
		dailyHrs, workHrs, maxBet     sql.NullInt64
	)
	err := db.QueryRow(`
		SELECT ticket_category_id, log_channel_id, staff_role_id,
		       auto_role_id, roles_channel_id, levelup_channel_id,
		       daily_cooldown_hours, work_cooldown_hours, max_bet
		FROM guild_settings WHERE guild_id=?`, guildID,
	).Scan(&ticketCat, &logChan, &staffRole,
		&autoRole, &rolesChan, &lvlupChan,
		&dailyHrs, &workHrs, &maxBet)
	if err == sql.ErrNoRows {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	cfg.TicketCategoryID = ticketCat.String
	cfg.LogChannelID = logChan.String
	cfg.StaffRoleID = staffRole.String
	cfg.AutoRoleID = autoRole.String
	cfg.RolesChannelID = rolesChan.String
	cfg.LevelUpChannelID = lvlupChan.String
	if dailyHrs.Valid && dailyHrs.Int64 > 0 {
		cfg.DailyCooldownHours = int(dailyHrs.Int64)
	}
	if workHrs.Valid && workHrs.Int64 > 0 {
		cfg.WorkCooldownHours = int(workHrs.Int64)
	}
	if maxBet.Valid {
		cfg.MaxBet = maxBet.Int64
	}
	return cfg, nil
}

// ── role_buttons ──────────────────────────────────────────────────────────────

func (db *DB) AddRoleButton(guildID, roleID, label, emoji string) error {
	_, err := db.Exec(`
		INSERT INTO role_buttons (guild_id, role_id, label, emoji)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(guild_id, role_id) DO UPDATE
			SET label = excluded.label,
			    emoji = excluded.emoji
	`, guildID, roleID, label, emoji)
	return err
}

func (db *DB) GetRoleButtons(guildID string) ([]RoleButton, error) {
	rows, err := db.Query(
		`SELECT id, guild_id, role_id, label, emoji FROM role_buttons WHERE guild_id=? ORDER BY id`,
		guildID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RoleButton
	for rows.Next() {
		var b RoleButton
		if err := rows.Scan(&b.ID, &b.GuildID, &b.RoleID, &b.Label, &b.Emoji); err != nil {
			continue
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
