package utils

import "github.com/bwmarrin/discordgo"

func Respond(s *discordgo.Session, i *discordgo.Interaction, content string) {
	s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: content},
	})
}

func RespondEmbed(s *discordgo.Session, i *discordgo.Interaction, embed *discordgo.MessageEmbed) {
	s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// RespondEphemeral envoie une réponse visible uniquement par l'auteur.
func RespondEphemeral(s *discordgo.Session, i *discordgo.Interaction, content string) {
	s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func RespondEmbedEphemeral(s *discordgo.Session, i *discordgo.Interaction, embed *discordgo.MessageEmbed) {
	s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}
