package events

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type CommandRegistrar interface {
	RegisterCommands()
}

func Ready(registrar CommandRegistrar) func(*discordgo.Session, *discordgo.Ready) {
	return func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("[bot] Connecté en tant que %s#%s", r.User.Username, r.User.Discriminator)
		registrar.RegisterCommands()
	}
}
