//go:build !cgo

package music

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

// playAudioFile est un stub quand CGO est désactivé (dev Windows sans gcc/libopus).
// La musique fonctionne pleinement en Docker Linux.
func playAudioFile(vc *discordgo.VoiceConnection, url string, done chan bool) {
	slog.Warn("audio non disponible sans CGO (gcc + libopus requis) — utilisez Docker pour la musique", "component", "music")
	done <- true
}
