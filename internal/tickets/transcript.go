package tickets

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Transcript struct {
	session   *discordgo.Session
	channelID string
}

func NewTranscript(s *discordgo.Session, channelID string) *Transcript {
	return &Transcript{session: s, channelID: channelID}
}

// Generate récupère les 100 derniers messages du salon et les formate en texte.
func (t *Transcript) Generate() (string, error) {
	messages, err := t.session.ChannelMessages(t.channelID, 100, "", "", "")
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	// Les messages arrivent du plus récent au plus ancien, on inverse
	for idx := len(messages) - 1; idx >= 0; idx-- {
		m := messages[idx]
		sb.WriteString(fmt.Sprintf(
			"[%s] %s#%s: %s\n",
			m.Timestamp.Format("2006-01-02 15:04:05"),
			m.Author.Username,
			m.Author.Discriminator,
			m.Content,
		))
	}

	return sb.String(), nil
}
