package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"

	"discord-bot/internal/commands"
	"discord-bot/internal/config"
	"discord-bot/internal/database"
	"discord-bot/internal/events"
	"discord-bot/internal/metrics"
)

const shutdownTimeout = 10 * time.Second

func main() {
	cfg := config.Load()

	db, err := database.New(cfg.DBPath)
	if err != nil {
		slog.Error("erreur base de données", "component", "main", "error", err)
		os.Exit(1)
	}
	defer db.Close() // 3. Fermé en dernier via defer

	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		slog.Error("erreur création session Discord", "component", "main", "error", err)
		os.Exit(1)
	}

	dg.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildBans

	handler := commands.NewHandler(dg, cfg, db)

	dg.AddHandler(events.Ready(handler))
	dg.AddHandler(events.InteractionCreate(handler))
	dg.AddHandler(events.VoiceStateUpdate(handler))
	dg.AddHandler(events.GuildMemberAdd(handler))
	dg.AddHandler(events.GuildBanAdd(handler))
	dg.AddHandler(events.GuildBanRemove(handler))
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		handler.HandleMessageCreate(s, m)
	})

	if err = dg.Open(); err != nil {
		slog.Error("erreur connexion Discord", "component", "main", "error", err)
		os.Exit(1)
	}

	healthSrv := startHealthServer(dg, cfg)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	<-ctx.Done()
	stop() // libère le signal handler immédiatement

	slog.Info("signal reçu, shutdown en cours", "component", "main")
	start := time.Now()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		if healthSrv != nil {
			_ = healthSrv.Shutdown(shutdownCtx) // 0. Fermer le health server en premier
		}
		dg.Close()         // 1. Fermer la session Discord
		handler.Shutdown() // 2. Arrêter les sub-managers (BJ, music, antispam)
		// 3. db.Close() est géré par defer
	}()

	select {
	case <-done:
		slog.Info("shutdown propre", "component", "main", "duration", time.Since(start).Round(time.Millisecond))
	case <-shutdownCtx.Done():
		slog.Warn("shutdown forcé (timeout 10s dépassé)", "component", "main", "duration", time.Since(start).Round(time.Millisecond))
	}
}

func startHealthServer(dg *discordgo.Session, cfg *config.Config) *http.Server {
	if cfg.HealthPort == 0 {
		return nil
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		snap := metrics.GetMetrics().Snapshot()
		dg.State.RLock()
		guilds := len(dg.State.Guilds)
		dg.State.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "ok",
			"uptime":  metrics.FormatUptime(snap.Uptime),
			"guilds":  guilds,
			"version": config.Version,
		})
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		dg.State.RLock()
		ready := dg.State.User != nil
		dg.State.RUnlock()
		if ready {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HealthPort),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("health server erreur", "component", "health", "error", err)
		}
	}()

	slog.Info("health server démarré", "component", "health", "port", cfg.HealthPort)
	return srv
}
