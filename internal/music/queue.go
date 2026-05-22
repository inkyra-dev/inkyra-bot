package music

import "sync"

type Track struct {
	Title     string
	URL       string // URL directe de l'audio (extraite par yt-dlp)
	Duration  string
	Requester string // username de celui qui a demandé
}

type Queue struct {
	mu     sync.Mutex
	tracks []Track
}

func (q *Queue) Add(t Track) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = append(q.tracks, t)
}

func (q *Queue) Pop() (Track, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.tracks) == 0 {
		return Track{}, false
	}
	t := q.tracks[0]
	q.tracks = q.tracks[1:]
	return t, true
}

func (q *Queue) Peek() (Track, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.tracks) == 0 {
		return Track{}, false
	}
	return q.tracks[0], true
}

func (q *Queue) List() []Track {
	q.mu.Lock()
	defer q.mu.Unlock()
	cp := make([]Track, len(q.tracks))
	copy(cp, q.tracks)
	return cp
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = nil
}

func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.tracks)
}
