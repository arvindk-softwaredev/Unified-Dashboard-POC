package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/unified-dashboard/backend/internal/cache"
	"github.com/unified-dashboard/backend/internal/github"
	"github.com/unified-dashboard/backend/internal/llm"
)

type Handler struct {
	github *github.Client
	llm    *llm.Client
	org    string
	log    *slog.Logger
	cache  *cache.Cache
}

func New(gh *github.Client, llmClient *llm.Client, org string, log *slog.Logger, cacheTTL time.Duration) *Handler {
	return &Handler{
		github: gh,
		llm:    llmClient,
		org:    org,
		log:    log,
		cache:  cache.New(cacheTTL),
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
