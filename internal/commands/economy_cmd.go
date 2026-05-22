package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/utils"
)

func (h *Handler) cmdBalance(s *discordgo.Session, i *discordgo.InteractionCreate) {
	targetID := i.Member.User.ID
	targetName := i.Member.User.Username

	if opts := i.ApplicationCommandData().Options; len(opts) > 0 {
		u := opts[0].UserValue(s)
		targetID = u.ID
		targetName = u.Username
	}

	u, err := h.econMgr.Balance(i.GuildID, targetID)
	if err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError("Erreur récupération solde."))
		return
	}

	embed := utils.EmbedFields(
		fmt.Sprintf("💳 Compte de %s", targetName), "", utils.ColorYellow,
		utils.Field("💰 Portefeuille", fmt.Sprintf("**%s** 🪙", coins(u.Wallet)), true),
		utils.Field("🏦 Banque", fmt.Sprintf("**%s** 🪙", coins(u.Bank)), true),
		utils.Field("💎 Total", fmt.Sprintf("**%s** 🪙", coins(u.Wallet+u.Bank)), false),
	)
	utils.RespondEmbed(s, i.Interaction, embed)
}

func (h *Handler) cmdDaily(s *discordgo.Session, i *discordgo.InteractionCreate) {
	result, err := h.econMgr.Daily(i.GuildID, i.Member.User.ID)
	if err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}

	streakStr := fmt.Sprintf("%d jour(s)", result.Streak)
	if result.Streak >= 7 {
		streakStr = "🔥 " + streakStr
	}

	embed := utils.EmbedFields(
		"🎁 Récompense quotidienne", "", utils.ColorGreen,
		utils.Field("💰 Reçu", fmt.Sprintf("**+%s** 🪙", coins(result.Reward)), true),
		utils.Field("🔥 Série", streakStr, true),
	)
	utils.RespondEmbed(s, i.Interaction, embed)
}

func (h *Handler) cmdWork(s *discordgo.Session, i *discordgo.InteractionCreate) {
	result, err := h.econMgr.Work(i.GuildID, i.Member.User.ID)
	if err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}

	embed := utils.EmbedFields(
		"💼 Travail", fmt.Sprintf("Tu as travaillé comme **%s** et tu as **%s**.", result.Job, result.Action),
		utils.ColorBlue,
		utils.Field("💰 Salaire", fmt.Sprintf("**+%s** 🪙", coins(result.Reward)), true),
	)
	utils.RespondEmbed(s, i.Interaction, embed)
}

func (h *Handler) cmdDeposit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	raw := i.ApplicationCommandData().Options[0].StringValue()
	var err error
	var amount int64

	if strings.ToLower(raw) == "all" {
		amount, err = h.econMgr.DepositAll(i.GuildID, i.Member.User.ID)
	} else {
		amount, err = parseCoins(raw)
		if err == nil {
			err = h.econMgr.Deposit(i.GuildID, i.Member.User.ID, amount)
		}
	}

	if err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}
	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("🏦 Dépôt", fmt.Sprintf("**%s** 🪙 déposés en banque.", coins(amount)), utils.ColorGreen))
}

func (h *Handler) cmdWithdraw(s *discordgo.Session, i *discordgo.InteractionCreate) {
	raw := i.ApplicationCommandData().Options[0].StringValue()
	var err error
	var amount int64

	if strings.ToLower(raw) == "all" {
		amount, err = h.econMgr.WithdrawAll(i.GuildID, i.Member.User.ID)
	} else {
		amount, err = parseCoins(raw)
		if err == nil {
			err = h.econMgr.Withdraw(i.GuildID, i.Member.User.ID, amount)
		}
	}

	if err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}
	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("💰 Retrait", fmt.Sprintf("**%s** 🪙 retirés de la banque.", coins(amount)), utils.ColorGreen))
}

func (h *Handler) cmdPay(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.ApplicationCommandData().Options
	target := opts[0].UserValue(s)
	amount := opts[1].IntValue()

	if err := h.econMgr.Pay(i.GuildID, i.Member.User.ID, target.ID, amount); err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}

	embed := utils.EmbedFields(
		"💸 Paiement", "", utils.ColorGreen,
		utils.Field("De", fmt.Sprintf("<@%s>", i.Member.User.ID), true),
		utils.Field("À", fmt.Sprintf("<@%s>", target.ID), true),
		utils.Field("Montant", fmt.Sprintf("**%s** 🪙", coins(amount)), false),
	)
	utils.RespondEmbed(s, i.Interaction, embed)
}

func (h *Handler) cmdShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	items, err := h.econMgr.Shop(i.GuildID)
	if err != nil || len(items) == 0 {
		utils.RespondEphemeral(s, i.Interaction, "La boutique est vide pour ce serveur.")
		return
	}

	var sb strings.Builder
	for _, it := range items {
		stock := "∞"
		if it.Stock >= 0 {
			stock = strconv.Itoa(it.Stock)
		}
		sb.WriteString(fmt.Sprintf(
			"`ID:%d` %s **%s** — %s 🪙 | revente: %s 🪙 | stock: %s\n",
			it.ID, it.Emoji, it.Name, coins(it.Price), coins(it.SellPrice), stock,
		))
	}
	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("🛒 Boutique", sb.String(), utils.ColorYellow))
}

func (h *Handler) cmdBuy(s *discordgo.Session, i *discordgo.InteractionCreate) {
	itemID := i.ApplicationCommandData().Options[0].IntValue()

	item, err := h.econMgr.Buy(i.GuildID, i.Member.User.ID, itemID)
	if err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}
	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("✅ Achat", fmt.Sprintf("Tu as acheté **%s %s** pour **%s** 🪙 !", item.Emoji, item.Name, coins(item.Price)), utils.ColorGreen))
}

func (h *Handler) cmdSell(s *discordgo.Session, i *discordgo.InteractionCreate) {
	itemID := i.ApplicationCommandData().Options[0].IntValue()

	item, err := h.econMgr.Sell(i.GuildID, i.Member.User.ID, itemID)
	if err != nil {
		utils.RespondEmbedEphemeral(s, i.Interaction, utils.EmbedError(err.Error()))
		return
	}
	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("💰 Vente", fmt.Sprintf("Tu as vendu **%s %s** pour **%s** 🪙 !", item.Emoji, item.Name, coins(item.SellPrice)), utils.ColorGreen))
}

func (h *Handler) cmdInventory(s *discordgo.Session, i *discordgo.InteractionCreate) {
	items, err := h.econMgr.Inventory(i.GuildID, i.Member.User.ID)
	if err != nil || len(items) == 0 {
		utils.RespondEphemeral(s, i.Interaction, "Ton inventaire est vide.")
		return
	}

	var sb strings.Builder
	for _, it := range items {
		sb.WriteString(fmt.Sprintf("`ID:%d` %s **%s** × %d\n", it.ID, it.Emoji, it.Name, it.Quantity))
	}
	utils.RespondEmbed(s, i.Interaction,
		utils.Embed("🎒 Inventaire", sb.String(), utils.ColorBlue))
}
