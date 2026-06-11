package metrics

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics holds all runtime counters. Thread-safe via atomic ops + RWMutex for the command map.
type Metrics struct {
	startTime    time.Time
	messages     atomic.Int64
	spamTimeouts atomic.Int64
	dbErrors     atomic.Int64

	cmdMu    sync.RWMutex
	commands map[string]*atomic.Int64
}

// Snapshot is a point-in-time read of all counters.
type Snapshot struct {
	Uptime        time.Duration
	Messages      int64
	SpamTimeouts  int64
	DBErrors      int64
	TotalCommands int64
	TopCommands   []CommandStat
}

// CommandStat pairs a command name with its invocation count.
type CommandStat struct {
	Name  string
	Count int64
}

var (
	instance *Metrics
	once     sync.Once
)

// GetMetrics returns the process-wide singleton, initialising it on first call.
func GetMetrics() *Metrics {
	once.Do(func() {
		instance = &Metrics{
			startTime: time.Now(),
			commands:  make(map[string]*atomic.Int64),
		}
	})
	return instance
}

// IncrCommand increments the per-command invocation counter.
func (m *Metrics) IncrCommand(name string) {
	m.cmdMu.RLock()
	c, ok := m.commands[name]
	m.cmdMu.RUnlock()

	if !ok {
		m.cmdMu.Lock()
		if c, ok = m.commands[name]; !ok {
			c = new(atomic.Int64)
			m.commands[name] = c
		}
		m.cmdMu.Unlock()
	}
	c.Add(1)
}

// IncrMessage increments the processed-message counter.
func (m *Metrics) IncrMessage() { m.messages.Add(1) }

// IncrSpamTimeout increments the anti-spam timeout counter.
func (m *Metrics) IncrSpamTimeout() { m.spamTimeouts.Add(1) }

// IncrDBError increments the database error counter.
func (m *Metrics) IncrDBError() { m.dbErrors.Add(1) }

// Snapshot returns a consistent read of all counters plus top-5 commands by usage.
func (m *Metrics) Snapshot() Snapshot {
	m.cmdMu.RLock()
	stats := make([]CommandStat, 0, len(m.commands))
	var total int64
	for name, c := range m.commands {
		n := c.Load()
		total += n
		stats = append(stats, CommandStat{Name: name, Count: n})
	}
	m.cmdMu.RUnlock()

	sort.Slice(stats, func(i, j int) bool { return stats[i].Count > stats[j].Count })
	if len(stats) > 5 {
		stats = stats[:5]
	}

	return Snapshot{
		Uptime:        time.Since(m.startTime),
		Messages:      m.messages.Load(),
		SpamTimeouts:  m.spamTimeouts.Load(),
		DBErrors:      m.dbErrors.Load(),
		TotalCommands: total,
		TopCommands:   stats,
	}
}

// FormatUptime formats a duration as "2j 4h 32m", "4h 05m", or "3m".
func FormatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60

	switch {
	case days > 0:
		return fmt.Sprintf("%dj %dh %02dm", days, hours, mins)
	case hours > 0:
		return fmt.Sprintf("%dh %02dm", hours, mins)
	default:
		return fmt.Sprintf("%dm", mins)
	}
}
