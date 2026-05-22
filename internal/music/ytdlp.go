package music

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type TrackInfo struct {
	Title    string  `json:"title"`
	URL      string  `json:"url"`      // URL directe du flux audio
	Duration float64 `json:"duration"` // en secondes
}

// GetInfo extrait le titre, l'URL audio et la durée via yt-dlp.
// query peut être une URL YouTube ou un terme de recherche.
func GetInfo(query string) (*TrackInfo, error) {
	// Si ce n'est pas une URL, préfixer avec ytsearch
	if !strings.HasPrefix(query, "http") {
		query = "ytsearch:" + query
	}

	cmd := exec.Command("yt-dlp",
		"--no-playlist",
		"-f", "bestaudio",
		"--print", `{"title":"%(title)s","url":"%(url)s","duration":%(duration)s}`,
		query,
	)
	cmd.Env = append(cmd.Environ(), "PYTHONUNBUFFERED=1")

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp: %w", err)
	}

	var info TrackInfo
	// On prend seulement la première ligne (en cas de playlist accidentelle)
	line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	if err := json.Unmarshal([]byte(line), &info); err != nil {
		return nil, fmt.Errorf("parsing yt-dlp output: %w", err)
	}

	return &info, nil
}

func formatDuration(seconds float64) string {
	d := time.Duration(seconds) * time.Second
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
