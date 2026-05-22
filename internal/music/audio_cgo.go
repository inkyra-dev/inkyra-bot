//go:build cgo

package music

import (
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/dgvoice"
)

func playAudioFile(vc *discordgo.VoiceConnection, url string, done chan bool) {
	dgvoice.PlayAudioFile(vc, url, done)
}
