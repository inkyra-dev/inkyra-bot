// Package games implémente les mini-jeux : coinflip, dice, slots.
// Le blackjack (multi-étapes avec boutons) est dans blackjack.go.
package games

import (
	"fmt"
	"math/rand"
	"strings"
)

// ── Coinflip ──────────────────────────────────────────────────────────────────

type CoinflipResult struct {
	Choice string // "pile" ou "face"
	Roll   string
	Won    bool
	Payout int64
}

func Coinflip(choice string, bet int64) *CoinflipResult {
	sides := []string{"pile", "face"}
	roll := sides[rand.Intn(2)]
	won := roll == choice
	payout := int64(0)
	if won {
		payout = bet * 2
	}
	return &CoinflipResult{Choice: choice, Roll: roll, Won: won, Payout: payout}
}

// ── Dice ──────────────────────────────────────────────────────────────────────

type DiceResult struct {
	Roll   int
	Won    bool
	Multi  float64 // multiplicateur de mise
	Payout int64
}

var diceEmojis = map[int]string{
	1: "⚀", 2: "⚁", 3: "⚂", 4: "⚃", 5: "⚄", 6: "⚅",
}

func (r *DiceResult) Emoji() string {
	return diceEmojis[r.Roll]
}

func Dice(bet int64) *DiceResult {
	roll := 1 + rand.Intn(6)
	var multi float64
	switch roll {
	case 6:
		multi = 3.0
	case 5:
		multi = 2.0
	case 4:
		multi = 1.5
	case 3:
		multi = 1.0 // remboursé
	default:
		multi = 0
	}
	payout := int64(float64(bet) * multi)
	return &DiceResult{Roll: roll, Won: multi >= 1.0, Multi: multi, Payout: payout}
}

// ── Slots ─────────────────────────────────────────────────────────────────────

type SlotsResult struct {
	Reels  [3]string
	Won    bool
	Multi  float64
	Payout int64
	Label  string
}

var slotSymbols = []struct {
	emoji string
	multi float64 // multiplicateur pour 3 identiques
	freq  int     // fréquence relative (plus élevé = plus fréquent)
}{
	{"🍎", 2.0, 6},
	{"🍊", 2.5, 5},
	{"🍋", 2.0, 6},
	{"🍇", 3.5, 4},
	{"⭐", 5.0, 3},
	{"💎", 10.0, 2},
	{"7️⃣", 20.0, 1},
}

func buildSlotPool() []string {
	pool := make([]string, 0, 32)
	for _, s := range slotSymbols {
		for i := 0; i < s.freq; i++ {
			pool = append(pool, s.emoji)
		}
	}
	return pool
}

var slotPool = buildSlotPool()

func Slots(bet int64) *SlotsResult {
	r := &SlotsResult{}
	for i := range r.Reels {
		r.Reels[i] = slotPool[rand.Intn(len(slotPool))]
	}

	if r.Reels[0] == r.Reels[1] && r.Reels[1] == r.Reels[2] {
		// 3 identiques
		for _, s := range slotSymbols {
			if s.emoji == r.Reels[0] {
				r.Multi = s.multi
				break
			}
		}
		r.Won = true
		r.Label = fmt.Sprintf("Jackpot ! × %.1f", r.Multi)
	} else if r.Reels[0] == r.Reels[1] || r.Reels[1] == r.Reels[2] || r.Reels[0] == r.Reels[2] {
		// 2 identiques : remboursement
		r.Multi = 1.0
		r.Won = true
		r.Label = "Deux identiques ! × 1.0"
	} else {
		r.Label = "Aucune combinaison"
	}

	r.Payout = int64(float64(bet) * r.Multi)
	return r
}

func (r *SlotsResult) Display() string {
	return strings.Join(r.Reels[:], " | ")
}
