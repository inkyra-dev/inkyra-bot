package events

import "github.com/bwmarrin/discordgo"

type InteractionHandler interface {
	Handle(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func InteractionCreate(h InteractionHandler) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		h.Handle(s, i)
	}
}
