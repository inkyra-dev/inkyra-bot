package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/music"
	"discord-bot/internal/utils"
)

// getUserVoiceChannel retourne l'ID du canal vocal de l'utilisateur sur ce serveur.
func getUserVoiceChannel(s *discordgo.Session, guildID, userID string) string {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return ""
	}
	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			return vs.ChannelID
		}
	}
	return ""
}

func (h *Handler) cmdPlay(s *discordgo.Session, i *discordgo.InteractionCreate) {
	voiceChannelID := getUserVoiceChannel(s, i.GuildID, i.Member.User.ID)
	if voiceChannelID == "" {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Tu dois être dans un salon vocal."))
		return
	}

	query := i.ApplicationCommandData().Options[0].StringValue()

	// ACK immédiat car yt-dlp peut prendre quelques secondes
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	info, err := music.GetInfo(query)
	if err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{utils.EmbedError("Impossible de trouver cette musique : " + err.Error())},
		})
		return
	}

	track := music.Track{
		Title:     info.Title,
		URL:       info.URL,
		Requester: i.Member.User.Username,
	}

	if err := h.music.Play(s, i.GuildID, voiceChannelID, track); err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{utils.EmbedError("Erreur lecture : " + err.Error())},
		})
		return
	}

	embed := utils.EmbedFields(
		"Ajouté à la queue", "", utils.ColorPurple,
		utils.Field("Titre", info.Title, false),
		utils.Field("Demandé par", i.Member.User.Username, true),
	)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

func (h *Handler) cmdSkip(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p, ok := h.music.Get(i.GuildID)
	if !ok || !p.IsPlaying() {
		utils.RespondEphemeral(s, i.Interaction, "Aucune musique en cours.")
		return
	}
	p.Skip()
	utils.RespondEmbed(s, i.Interaction, utils.Embed("Skip", "Musique suivante ⏭", utils.ColorYellow))
}

func (h *Handler) cmdStop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p, ok := h.music.Get(i.GuildID)
	if !ok {
		utils.RespondEphemeral(s, i.Interaction, "Aucune musique en cours.")
		return
	}
	p.Stop()
	utils.RespondEmbed(s, i.Interaction, utils.Embed("Stop", "Lecture arrêtée, queue vidée. ⏹", utils.ColorRed))
}

func (h *Handler) cmdQueue(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p, ok := h.music.Get(i.GuildID)
	if !ok {
		utils.RespondEphemeral(s, i.Interaction, "La queue est vide.")
		return
	}

	current := p.Current()
	tracks := p.Queue().List()

	var sb strings.Builder
	if current != nil {
		sb.WriteString(fmt.Sprintf("**En cours :** %s\n\n", current.Title))
	}

	if len(tracks) == 0 {
		sb.WriteString("*Queue vide*")
	} else {
		for idx, t := range tracks {
			sb.WriteString(fmt.Sprintf("`%d.` %s — demandé par %s\n", idx+1, t.Title, t.Requester))
		}
	}

	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("File d'attente", sb.String(), utils.ColorPurple),
	)
}

func (h *Handler) cmdPause(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p, ok := h.music.Get(i.GuildID)
	if !ok || !p.IsPlaying() {
		utils.RespondEphemeral(s, i.Interaction, "Aucune musique en cours.")
		return
	}
	p.Pause()
	utils.RespondEmbed(s, i.Interaction, utils.Embed("Pause", "Lecture mise en pause ⏸", utils.ColorYellow))
}

func (h *Handler) cmdResume(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p, ok := h.music.Get(i.GuildID)
	if !ok || !p.IsPlaying() {
		utils.RespondEphemeral(s, i.Interaction, "Aucune musique en cours.")
		return
	}
	p.Resume()
	utils.RespondEmbed(s, i.Interaction, utils.Embed("Reprise", "Lecture reprise ▶️", utils.ColorGreen))
}

func (h *Handler) cmdLeave(s *discordgo.Session, i *discordgo.InteractionCreate) {
	p, ok := h.music.Get(i.GuildID)
	if !ok {
		utils.RespondEphemeral(s, i.Interaction, "Je ne suis dans aucun salon vocal.")
		return
	}
	p.Stop()
	utils.RespondEmbed(s, i.Interaction, utils.Embed("Déconnexion", "À bientôt 👋", utils.ColorGrey))
}
