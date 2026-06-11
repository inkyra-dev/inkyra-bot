package events

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

type CommandRegistrar interface {
	RegisterCommands()
}

func Ready(registrar CommandRegistrar) func(*discordgo.Session, *discordgo.Ready) {
	return func(s *discordgo.Session, r *discordgo.Ready) {
		slog.Info("connecté", "component", "bot", "username", r.User.Username, "discriminator", r.User.Discriminator)
		registrar.RegisterCommands()
	}
}
