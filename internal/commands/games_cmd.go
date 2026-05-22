package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/games"
	"discord-bot/internal/utils"
)

// ── Coinflip ──────────────────────────────────────────────────────────────────

func (h *Handler) cmdCoinflip(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.ApplicationCommandData().Options
	choice := opts[0].StringValue()
	bet := opts[1].IntValue()

	if bet <= 0 {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("La mise doit être > 0."))
		return
	}
	if err := h.gamesMgr.Debit(i.GuildID, i.Member.User.ID, bet); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}

	result := games.Coinflip(choice, bet)

	color := utils.ColorRed
	status := "❌ Perdu"
	if result.Won {
		color = utils.ColorGreen
		status = "✅ Gagné"
		h.gamesMgr.Credit(i.GuildID, i.Member.User.ID, result.Payout)
	}

	embed := utils.EmbedFields(
		fmt.Sprintf("🪙 Coinflip — %s", status), "", color,
		utils.Field("Ton choix", result.Choice, true),
		utils.Field("Résultat", result.Roll, true),
		utils.Field("Mise", fmt.Sprintf("%s 🪙", coins(bet)), true),
		utils.Field("Gain/Perte", func() string {
			if result.Won {
				return fmt.Sprintf("+%s 🪙", coins(result.Payout-bet))
			}
			return fmt.Sprintf("-%s 🪙", coins(bet))
		}(), true),
	)
	utils.RespondEmbed(s, i.Interaction, embed)
}

// ── Dice ──────────────────────────────────────────────────────────────────────

func (h *Handler) cmdDice(s *discordgo.Session, i *discordgo.InteractionCreate) {
	bet := i.ApplicationCommandData().Options[0].IntValue()

	if bet <= 0 {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("La mise doit être > 0."))
		return
	}
	if err := h.gamesMgr.Debit(i.GuildID, i.Member.User.ID, bet); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}

	result := games.Dice(bet)
	net := result.Payout - bet

	if result.Payout > 0 {
		h.gamesMgr.Credit(i.GuildID, i.Member.User.ID, result.Payout)
	}

	color := utils.ColorRed
	if result.Won {
		color = utils.ColorGreen
	}

	embed := utils.EmbedFields(
		fmt.Sprintf("%s Dé — ×%.1f", result.Emoji(), result.Multi), "", color,
		utils.Field("Résultat", fmt.Sprintf("**%d** %s", result.Roll, result.Emoji()), true),
		utils.Field("Mise", fmt.Sprintf("%s 🪙", coins(bet)), true),
		utils.Field("Gain/Perte", func() string {
			if net >= 0 {
				return fmt.Sprintf("+%s 🪙", coins(net))
			}
			return fmt.Sprintf("-%s 🪙", coins(-net))
		}(), true),
	)
	utils.RespondEmbed(s, i.Interaction, embed)
}

// ── Slots ─────────────────────────────────────────────────────────────────────

func (h *Handler) cmdSlots(s *discordgo.Session, i *discordgo.InteractionCreate) {
	bet := i.ApplicationCommandData().Options[0].IntValue()

	if bet <= 0 {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("La mise doit être > 0."))
		return
	}
	if err := h.gamesMgr.Debit(i.GuildID, i.Member.User.ID, bet); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}

	result := games.Slots(bet)
	net := result.Payout - bet

	if result.Payout > 0 {
		h.gamesMgr.Credit(i.GuildID, i.Member.User.ID, result.Payout)
	}

	color := utils.ColorRed
	if result.Won {
		color = utils.ColorGreen
	}

	embed := utils.EmbedFields(
		"🎰 Machines à sous",
		fmt.Sprintf("## %s\n**%s**", result.Display(), result.Label),
		color,
		utils.Field("Mise", fmt.Sprintf("%s 🪙", coins(bet)), true),
		utils.Field("Gain/Perte", func() string {
			if net >= 0 {
				return fmt.Sprintf("+%s 🪙", coins(net))
			}
			return fmt.Sprintf("-%s 🪙", coins(-net))
		}(), true),
	)
	utils.RespondEmbed(s, i.Interaction, embed)
}

// ── Blackjack ─────────────────────────────────────────────────────────────────

func (h *Handler) cmdBlackjack(s *discordgo.Session, i *discordgo.InteractionCreate) {
	bet := i.ApplicationCommandData().Options[0].IntValue()

	if bet <= 0 {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("La mise doit être > 0."))
		return
	}
	if h.gamesMgr.BJ.HasActive(i.GuildID, i.Member.User.ID) {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Tu as déjà une partie en cours ! Utilise Hit ou Stand."))
		return
	}
	if err := h.gamesMgr.Debit(i.GuildID, i.Member.User.ID, bet); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}

	session, outcome := h.gamesMgr.BJ.Start(i.GuildID, i.Member.User.ID, i.ChannelID, bet)

	if outcome == games.BJBlackjack {
		payout := session.Payout(games.BJBlackjack)
		h.gamesMgr.Credit(i.GuildID, i.Member.User.ID, bet+payout)
		h.gamesMgr.BJ.Delete(i.GuildID, i.Member.User.ID)
		utils.RespondEmbed(s, i.Interaction, utils.EmbedFields(
			"🃏 Blackjack — Blackjack naturel ! 🎉", "", utils.ColorGreen,
			utils.Field("Ta main", session.PlayerHandStr(false), false),
			utils.Field("Gain", fmt.Sprintf("+%s 🪙 (×1.5)", coins(payout)), true),
		))
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{bjEmbed(session, false)},
			Components: bjButtons(),
		},
	})

	msg, _ := s.InteractionResponse(i.Interaction)
	if msg != nil {
		session.MessageID = msg.ID
	}
}

// Handlers boutons Hit / Stand

func (h *Handler) bjHit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	session, ok := h.gamesMgr.BJ.Get(i.GuildID, i.Member.User.ID)
	if !ok || session.Done {
		utils.RespondEphemeral(s, i.Interaction, "Aucune partie en cours.")
		return
	}

	outcome := session.Hit()
	h.bjRespond(s, i, session, outcome)
}

func (h *Handler) bjStand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	session, ok := h.gamesMgr.BJ.Get(i.GuildID, i.Member.User.ID)
	if !ok || session.Done {
		utils.RespondEphemeral(s, i.Interaction, "Aucune partie en cours.")
		return
	}

	outcome := session.Stand()
	h.bjRespond(s, i, session, outcome)
}

func (h *Handler) bjRespond(s *discordgo.Session, i *discordgo.InteractionCreate, session *games.BlackjackSession, outcome games.BJOutcome) {
	if outcome == games.BJOngoing {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{bjEmbed(session, false)},
				Components: bjButtons(),
			},
		})
		return
	}

	// Partie terminée
	h.gamesMgr.BJ.Delete(i.GuildID, i.Member.User.ID)
	payout := session.Payout(outcome)

	switch outcome {
	case games.BJPlayerWin:
		h.gamesMgr.Credit(i.GuildID, i.Member.User.ID, session.Bet+payout)
	case games.BJPush:
		h.gamesMgr.Refund(i.GuildID, i.Member.User.ID, session.Bet)
	// BJBust et BJDealerWin : mise déjà déduite, rien à rembourser
	}

	title, color := bjOutcomeLabel(outcome)
	embed := bjEmbed(session, true)
	embed.Title = title
	embed.Color = color

	gainStr := "0 🪙"
	if payout > 0 {
		gainStr = fmt.Sprintf("+%s 🪙", coins(payout))
	} else if payout < 0 {
		gainStr = fmt.Sprintf("-%s 🪙", coins(session.Bet))
	}
	embed.Fields = append(embed.Fields, utils.Field("Résultat", gainStr, true))

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{},
		},
	})
}

func bjEmbed(s *games.BlackjackSession, reveal bool) *discordgo.MessageEmbed {
	return utils.EmbedFields(
		"🃏 Blackjack", fmt.Sprintf("Mise : **%s** 🪙", coins(s.Bet)), utils.ColorBlue,
		utils.Field("Ta main", s.PlayerHandStr(false), false),
		utils.Field("Main du dealer", s.DealerHandStr(reveal), false),
	)
}

func bjButtons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{Components: []discordgo.MessageComponent{
			discordgo.Button{Label: "🃏 Hit", Style: discordgo.PrimaryButton, CustomID: "bj_hit"},
			discordgo.Button{Label: "✋ Stand", Style: discordgo.SecondaryButton, CustomID: "bj_stand"},
		}},
	}
}

func bjOutcomeLabel(outcome games.BJOutcome) (string, int) {
	switch outcome {
	case games.BJPlayerWin:
		return "🃏 Blackjack — Victoire ! 🎉", utils.ColorGreen
	case games.BJDealerWin:
		return "🃏 Blackjack — Le dealer gagne", utils.ColorRed
	case games.BJBust:
		return "🃏 Blackjack — Bust ! (> 21)", utils.ColorRed
	case games.BJPush:
		return "🃏 Blackjack — Égalité", utils.ColorYellow
	default:
		return "🃏 Blackjack", utils.ColorBlue
	}
}
