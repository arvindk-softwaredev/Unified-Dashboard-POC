package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	ghclient "github.com/unified-dashboard/backend/internal/github"
)

func (h *Handler) RepoInsights(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	name := chi.URLParam(r, "name")
	if owner == "" || name == "" {
		writeError(w, http.StatusBadRequest, "owner and repository name are required")
		return
	}

	key := "insights:" + owner + "/" + name

	insights, hit, err := getOrFetch(h, r, key, func() (*ghclient.RepoInsights, error) {
		ctx, cancel := contextWithTimeout(r, 90*time.Second)
		defer cancel()
		return h.github.RepoInsights(ctx, owner, name)
	})
	if err != nil {
		h.log.Error("failed repo insights", "owner", owner, "repo", name, "error", err)
		writeError(w, http.StatusBadGateway, "failed to fetch repository insights from GitHub")
		return
	}
	if insights.Categories == nil {
		insights.Categories = []ghclient.CategoryInsight{}
	}

	h.writeCached(w, r, key, hit, insights)
}
