package repositories

import (
	"fmt"

	"discord-bot/internal/database"
)

type Item struct {
	ID          int64
	GuildID     string
	Name        string
	Description string
	Price       int64
	SellPrice   int64
	Emoji       string
	Stock       int
}

type InventoryItem struct {
	Item
	Quantity int
}

type ItemsRepo struct {
	db *database.DB
}

func NewItemsRepo(db *database.DB) *ItemsRepo { return &ItemsRepo{db: db} }

func (r *ItemsRepo) GetAll(guildID string) ([]Item, error) {
	rows, err := r.db.Query(`
		SELECT id, guild_id, name, description, price, sell_price, emoji, stock
		FROM items WHERE guild_id=? ORDER BY price ASC
	`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Item
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.GuildID, &it.Name, &it.Description, &it.Price, &it.SellPrice, &it.Emoji, &it.Stock); err != nil {
			continue
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *ItemsRepo) GetByID(guildID string, id int64) (*Item, error) {
	it := &Item{}
	err := r.db.QueryRow(`
		SELECT id, guild_id, name, description, price, sell_price, emoji, stock
		FROM items WHERE id=? AND guild_id=?
	`, id, guildID).Scan(&it.ID, &it.GuildID, &it.Name, &it.Description, &it.Price, &it.SellPrice, &it.Emoji, &it.Stock)
	if err != nil {
		return nil, fmt.Errorf("item introuvable")
	}
	return it, nil
}

func (r *ItemsRepo) Buy(guildID, userID string, itemID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var price, stock int64
	if err := tx.QueryRow(`SELECT price, stock FROM items WHERE id=? AND guild_id=?`, itemID, guildID).Scan(&price, &stock); err != nil {
		return fmt.Errorf("item introuvable")
	}
	if stock == 0 {
		return fmt.Errorf("rupture de stock")
	}

	var wallet int64
	tx.QueryRow(`SELECT COALESCE(wallet,0) FROM economy WHERE guild_id=? AND user_id=?`, guildID, userID).Scan(&wallet)
	if wallet < price {
		return fmt.Errorf("solde insuffisant (%d 🪙 requis, %d 🪙 disponible)", price, wallet)
	}

	if _, err = tx.Exec(`UPDATE economy SET wallet=wallet-? WHERE guild_id=? AND user_id=?`, price, guildID, userID); err != nil {
		return err
	}
	if _, err = tx.Exec(`
		INSERT INTO inventory (user_id, guild_id, item_id, quantity)
		VALUES (?,?,?,1)
		ON CONFLICT(user_id, guild_id, item_id) DO UPDATE SET quantity=quantity+1
	`, userID, guildID, itemID); err != nil {
		return err
	}
	if stock > 0 {
		if _, err = tx.Exec(`UPDATE items SET stock=stock-1 WHERE id=?`, itemID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *ItemsRepo) Sell(guildID, userID string, itemID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var sellPrice int64
	if err := tx.QueryRow(`SELECT sell_price FROM items WHERE id=?`, itemID).Scan(&sellPrice); err != nil {
		return fmt.Errorf("item introuvable")
	}
	if sellPrice == 0 {
		return fmt.Errorf("cet item ne peut pas être revendu")
	}

	var qty int
	if err := tx.QueryRow(`SELECT quantity FROM inventory WHERE user_id=? AND guild_id=? AND item_id=?`, userID, guildID, itemID).Scan(&qty); err != nil || qty == 0 {
		return fmt.Errorf("tu ne possèdes pas cet item")
	}

	if qty == 1 {
		_, err = tx.Exec(`DELETE FROM inventory WHERE user_id=? AND guild_id=? AND item_id=?`, userID, guildID, itemID)
	} else {
		_, err = tx.Exec(`UPDATE inventory SET quantity=quantity-1 WHERE user_id=? AND guild_id=? AND item_id=?`, userID, guildID, itemID)
	}
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO economy (user_id, guild_id, wallet, bank) VALUES (?,?,?,100)
		ON CONFLICT(user_id, guild_id) DO UPDATE SET wallet=wallet+excluded.wallet
	`, userID, guildID, sellPrice)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ItemsRepo) GetInventory(guildID, userID string) ([]InventoryItem, error) {
	rows, err := r.db.Query(`
		SELECT i.id, i.guild_id, i.name, i.description, i.price, i.sell_price, i.emoji, i.stock, inv.quantity
		FROM inventory inv
		JOIN items i ON i.id=inv.item_id
		WHERE inv.user_id=? AND inv.guild_id=?
		ORDER BY i.name
	`, userID, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []InventoryItem
	for rows.Next() {
		var it InventoryItem
		if err := rows.Scan(&it.ID, &it.GuildID, &it.Name, &it.Description, &it.Price, &it.SellPrice, &it.Emoji, &it.Stock, &it.Quantity); err != nil {
			continue
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *ItemsRepo) Create(it Item) (int64, error) {
	res, err := r.db.Exec(`
		INSERT INTO items (guild_id, name, description, price, sell_price, emoji, stock)
		VALUES (?,?,?,?,?,?,?)
	`, it.GuildID, it.Name, it.Description, it.Price, it.SellPrice, it.Emoji, it.Stock)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
