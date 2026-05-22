package events

import (
	"github.com/bwmarrin/discordgo"
)

type VoiceCleaner interface {
	HandleVoiceStateUpdate(s *discordgo.Session, vs *discordgo.VoiceStateUpdate)
}

func VoiceStateUpdate(cleaner VoiceCleaner) func(*discordgo.Session, *discordgo.VoiceStateUpdate) {
	return func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
		cleaner.HandleVoiceStateUpdate(s, vs)
	}
}
