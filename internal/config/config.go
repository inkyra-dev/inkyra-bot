package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"

	"discord-bot/internal/logger"
)

// Version est la version courante du bot, affichée dans /health et /stats.
const Version = "1.0.3"

type Config struct {
	Token            string
	GuildID          string
	StaffRoleID      string
	TicketCategoryID string
	LogChannelID     string
	DBPath           string
	LogFormat        string
	LogLevel         string
	HealthPort       int
}

var validLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

func Load() *Config {
	_ = godotenv.Load()
	logger.Init()

	// ── Champs obligatoires ─────────────────────────────────────────────────────
	token := os.Getenv("TOKEN")
	if token == "" {
		slog.Error("TOKEN manquant dans .env", "component", "config")
		os.Exit(1)
	}

	guildID := os.Getenv("GUILD_ID")
	if guildID == "" {
		slog.Error("GUILD_ID manquant dans .env", "component", "config")
		os.Exit(1)
	}

	// ── DB_PATH ─────────────────────────────────────────────────────────────────
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/bot.db"
	}
	validateDBPath(dbPath)

	// ── LOG_LEVEL ───────────────────────────────────────────────────────────────
	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if logLevel != "" && !validLogLevels[logLevel] {
		slog.Warn("LOG_LEVEL invalide, fallback info", "component", "config", "valeur", logLevel)
		logLevel = "info"
	}

	// ── HEALTH_PORT ─────────────────────────────────────────────────────────────
	healthPort := 8080
	if portStr := os.Getenv("HEALTH_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			healthPort = p
		} else {
			slog.Warn("HEALTH_PORT invalide, fallback 8080", "component", "config", "valeur", portStr)
		}
	}

	return &Config{
		Token:            token,
		GuildID:          guildID,
		StaffRoleID:      os.Getenv("STAFF_ROLE_ID"),
		TicketCategoryID: os.Getenv("TICKET_CATEGORY_ID"),
		LogChannelID:     os.Getenv("LOG_CHANNEL_ID"),
		DBPath:           dbPath,
		LogFormat:        os.Getenv("LOG_FORMAT"),
		LogLevel:         logLevel,
		HealthPort:       healthPort,
	}
}

// validateDBPath émet un warning si le répertoire parent n'existe pas ou n'est pas accessible en écriture.
func validateDBPath(dbPath string) {
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		slog.Warn("répertoire parent de DB_PATH inexistant (sera créé au démarrage)", "component", "config", "path", dir)
		return
	}
	f, err := os.CreateTemp(dir, ".write-check-*")
	if err != nil {
		slog.Warn("répertoire parent de DB_PATH non accessible en écriture", "component", "config", "path", dir, "error", err)
		return
	}
	f.Close()
	os.Remove(f.Name())
}
