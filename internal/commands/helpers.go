package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// coins formate un nombre de coins avec séparateur de milliers.
func coins(n int64) string {
	s := strconv.FormatInt(n, 10)
	if n < 0 {
		s = s[1:]
	}
	var sb strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			sb.WriteByte(',')
		}
		sb.WriteRune(c)
	}
	if n < 0 {
		return "-" + sb.String()
	}
	return sb.String()
}

// parseCoins convertit une chaîne en int64 (accepte "1000" ou "1,000").
func parseCoins(s string) (int64, error) {
	s = strings.ReplaceAll(s, ",", "")
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("montant invalide : `%s`", s)
	}
	return n, nil
}

// usernameOrID retourne le pseudo Discord ou l'ID si introuvable.
func usernameOrID(s *discordgo.Session, guildID, userID string) string {
	if m, err := s.State.Member(guildID, userID); err == nil {
		name := m.Nick
		if name == "" {
			name = m.User.Username
		}
		return name
	}
	return userID
}

// leaderboardButtons génère les boutons Précédent / Suivant pour un leaderboard.
// prefix est un identifiant court ("xplb" ou "ecolb") pour distinguer les types.
func leaderboardButtons(page, totalPages int, prefix string) []discordgo.MessageComponent {
	if totalPages <= 1 {
		return nil
	}
	row := discordgo.ActionsRow{}
	if page > 1 {
		row.Components = append(row.Components, discordgo.Button{
			Label:    "◀ Précédent",
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("%s:%d", prefix, page-1),
		})
	}
	if page < totalPages {
		row.Components = append(row.Components, discordgo.Button{
			Label:    "Suivant ▶",
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("%s:%d", prefix, page+1),
		})
	}
	if len(row.Components) == 0 {
		return nil
	}
	return []discordgo.MessageComponent{row}
}
