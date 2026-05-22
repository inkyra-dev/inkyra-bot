package utils

import "github.com/bwmarrin/discordgo"

const (
	ColorGreen  = 0x2ecc71
	ColorRed    = 0xe74c3c
	ColorBlue   = 0x3498db
	ColorYellow = 0xf39c12
	ColorPurple = 0x9b59b6
	ColorGrey   = 0x95a5a6
)

func Embed(title, description string, color int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       color,
	}
}

func EmbedFields(title, description string, color int, fields ...*discordgo.MessageEmbedField) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       color,
		Fields:      fields,
	}
}

func Field(name, value string, inline bool) *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{Name: name, Value: value, Inline: inline}
}

func EmbedError(description string) *discordgo.MessageEmbed {
	return Embed("Erreur", description, ColorRed)
}
