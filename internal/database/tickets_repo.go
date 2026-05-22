package database

import (
	"database/sql"
	"time"
)

type Ticket struct {
	ID        int64
	ChannelID string
	GuildID   string
	UserID    string
	Status    string
	CreatedAt time.Time
	ClosedAt  *time.Time
}

func (db *DB) CreateTicket(channelID, guildID, userID string) error {
	_, err := db.Exec(
		`INSERT INTO tickets (channel_id, guild_id, user_id) VALUES (?, ?, ?)`,
		channelID, guildID, userID,
	)
	return err
}

func (db *DB) GetOpenTicketByUser(guildID, userID string) (*Ticket, error) {
	row := db.QueryRow(
		`SELECT id, channel_id, guild_id, user_id, status, created_at
		 FROM tickets WHERE guild_id=? AND user_id=? AND status='open'`,
		guildID, userID,
	)
	t := &Ticket{}
	err := row.Scan(&t.ID, &t.ChannelID, &t.GuildID, &t.UserID, &t.Status, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (db *DB) GetTicketByChannel(channelID string) (*Ticket, error) {
	row := db.QueryRow(
		`SELECT id, channel_id, guild_id, user_id, status, created_at
		 FROM tickets WHERE channel_id=?`,
		channelID,
	)
	t := &Ticket{}
	err := row.Scan(&t.ID, &t.ChannelID, &t.GuildID, &t.UserID, &t.Status, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (db *DB) CloseTicket(channelID string) error {
	_, err := db.Exec(
		`UPDATE tickets SET status='closed', closed_at=? WHERE channel_id=?`,
		time.Now(), channelID,
	)
	return err
}

func (db *DB) IsOpenTicket(channelID string) bool {
	var count int
	db.QueryRow(
		`SELECT COUNT(*) FROM tickets WHERE channel_id=? AND status='open'`,
		channelID,
	).Scan(&count)
	return count > 0
}

// ErrNoTicket is returned when a query finds no matching ticket.
var ErrNoTicket = sql.ErrNoRows
