package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Token            string
	GuildID          string
	StaffRoleID      string
	TicketCategoryID string
	LogChannelID     string
	DBPath           string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("[config] Pas de .env trouvé, utilisation des variables d'environnement")
	}

	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("[config] TOKEN manquant")
	}

	guildID := os.Getenv("GUILD_ID")
	if guildID == "" {
		log.Fatal("[config] GUILD_ID manquant")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/bot.db"
	}

	return &Config{
		Token:            token,
		GuildID:          guildID,
		StaffRoleID:      os.Getenv("STAFF_ROLE_ID"),
		TicketCategoryID: os.Getenv("TICKET_CATEGORY_ID"),
		LogChannelID:     os.Getenv("LOG_CHANNEL_ID"),
		DBPath:           dbPath,
	}
}
