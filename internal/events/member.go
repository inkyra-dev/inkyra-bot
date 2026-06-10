package events

import "github.com/bwmarrin/discordgo"

type MemberJoinHandler interface {
	HandleMemberJoin(s *discordgo.Session, guildID, userID string)
}

func GuildMemberAdd(h MemberJoinHandler) func(*discordgo.Session, *discordgo.GuildMemberAdd) {
	return func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
		if m.Member == nil || m.Member.User == nil {
			return
		}
		h.HandleMemberJoin(s, m.GuildID, m.Member.User.ID)
	}
}
