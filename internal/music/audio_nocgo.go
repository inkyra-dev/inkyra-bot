//go:build !cgo

package music

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// playAudioFile est un stub quand CGO est désactivé (dev Windows sans gcc/libopus).
// La musique fonctionne pleinement en Docker Linux.
func playAudioFile(vc *discordgo.VoiceConnection, url string, done chan bool) {
	log.Println("[music] Audio non disponible sans CGO (gcc + libopus requis). Utilisez Docker pour la musique.")
	done <- true
}
