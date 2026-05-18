package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/unified-dashboard/backend/internal/config"
	ghclient "github.com/unified-dashboard/backend/internal/github"
	"github.com/unified-dashboard/backend/internal/handler"
	"github.com/unified-dashboard/backend/internal/server"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	cfg := config.Load()

	if !cfg.GitHubAuthenticated() {
		log.Warn("GITHUB_TOKEN is not set — using unauthenticated GitHub API (60 requests/hour)",
			"hint", "create backend/.env with GITHUB_TOKEN=ghp_... (do not edit .env.example)")
	} else {
		log.Info("GITHUB_TOKEN loaded — using authenticated GitHub API",
			"token_prefix", cfg.GitHubToken[:min(7, len(cfg.GitHubToken))]+"...")
	}

	gh := ghclient.NewClient(cfg.GitHubToken)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	gh.LogRateLimit(ctx)

	repos, err := gh.ListOrgRepositories(ctx, cfg.GitHubOrg)
	if err != nil {
		log.Error("startup: failed to fetch repositories", "org", cfg.GitHubOrg, "error", err)
		os.Exit(1)
	}

	log.Info("repositories loaded", "org", cfg.GitHubOrg, "count", len(repos))

	log.Info("API cache enabled", "ttl_minutes", cfg.CacheTTL.Minutes())

	h := handler.New(gh, cfg.GitHubOrg, log, cfg.CacheTTL)
	srv := server.New(cfg, h)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      srv,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("server started", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", "error", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
