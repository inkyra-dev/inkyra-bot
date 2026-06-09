package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/utils"
)

func (h *Handler) cmdAchievements(s *discordgo.Session, i *discordgo.InteractionCreate) {
	targetID := i.Member.User.ID
	targetName := i.Member.User.Username

	if opts := i.ApplicationCommandData().Options; len(opts) > 0 {
		u := opts[0].UserValue(s)
		targetID = u.ID
		targetName = u.Username
	}

	statuses, err := h.achMgr.GetStatus(i.GuildID, targetID)
	if err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Erreur récupération achievements."))
		return
	}

	unlocked := 0
	var sb strings.Builder
	for _, st := range statuses {
		if st.Unlocked {
			unlocked++
			fmt.Fprintf(&sb, "✅ %s **%s** — %s\n", st.Emoji, st.Name, st.Description)
		} else {
			fmt.Fprintf(&sb, "🔒 %s ~~%s~~ — %s\n", st.Emoji, st.Name, st.Description)
		}
	}

	embed := utils.EmbedFields(
		fmt.Sprintf("🏆 Achievements de %s", targetName),
		sb.String(),
		utils.ColorYellow,
		utils.Field("Progression", fmt.Sprintf("**%d / %d** débloqués", unlocked, len(statuses)), false),
	)
	utils.RespondEmbed(s, i.Interaction, embed)
}
