package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/commands"
	"discord-bot/internal/config"
	"discord-bot/internal/database"
	"discord-bot/internal/events"
)

func main() {
	cfg := config.Load()

	db, err := database.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("[main] Erreur base de données: %v", err)
	}
	defer db.Close()

	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		log.Fatalf("[main] Erreur création session: %v", err)
	}

	dg.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildBans

	handler := commands.NewHandler(dg, cfg, db)

	dg.AddHandler(events.Ready(handler))
	dg.AddHandler(events.InteractionCreate(handler))
	dg.AddHandler(events.VoiceStateUpdate(handler))
	dg.AddHandler(events.GuildMemberAdd(handler))
	dg.AddHandler(events.GuildBanAdd(handler))
	dg.AddHandler(events.GuildBanRemove(handler))
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		handler.HandleMessageCreate(s, m)
	})

	if err = dg.Open(); err != nil {
		log.Fatalf("[main] Erreur connexion: %v", err)
	}
	defer dg.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	handler.Shutdown()
	log.Println("[main] Arrêt propre...")
}
