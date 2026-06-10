package events

import "github.com/bwmarrin/discordgo"

type BanHandler interface {
	HandleBanAdd(s *discordgo.Session, guildID, userID string)
	HandleBanRemove(s *discordgo.Session, guildID, userID string)
}

func GuildBanAdd(h BanHandler) func(*discordgo.Session, *discordgo.GuildBanAdd) {
	return func(s *discordgo.Session, b *discordgo.GuildBanAdd) {
		if b.User == nil {
			return
		}
		h.HandleBanAdd(s, b.GuildID, b.User.ID)
	}
}

func GuildBanRemove(h BanHandler) func(*discordgo.Session, *discordgo.GuildBanRemove) {
	return func(s *discordgo.Session, b *discordgo.GuildBanRemove) {
		if b.User == nil {
			return
		}
		h.HandleBanRemove(s, b.GuildID, b.User.ID)
	}
}
