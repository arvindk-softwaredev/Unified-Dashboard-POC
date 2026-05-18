package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	ghclient "github.com/unified-dashboard/backend/internal/github"
)

type goodFirstIssuesResponse struct {
	Repository string           `json:"repository"`
	Count      int              `json:"count"`
	Issues     []ghclient.Issue `json:"issues"`
}

func (h *Handler) ListGoodFirstIssues(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	name := chi.URLParam(r, "name")
	if owner == "" || name == "" {
		writeError(w, http.StatusBadRequest, "owner and repository name are required")
		return
	}

	key := "good-first-issues:" + owner + "/" + name

	resp, hit, err := getOrFetch(h, r, key, func() (goodFirstIssuesResponse, error) {
		ctx, cancel := contextWithTimeout(r, 45*time.Second)
		defer cancel()

		issues, err := h.github.ListGoodFirstIssues(ctx, owner, name)
		if err != nil {
			return goodFirstIssuesResponse{}, err
		}
		if issues == nil {
			issues = []ghclient.Issue{}
		}
		return goodFirstIssuesResponse{
			Repository: owner + "/" + name,
			Count:      len(issues),
			Issues:     issues,
		}, nil
	})
	if err != nil {
		h.log.Error("failed to list good first issues", "owner", owner, "repo", name, "error", err)
		writeError(w, http.StatusBadGateway, "failed to fetch issues from GitHub")
		return
	}

	h.writeCached(w, r, key, hit, resp)
}
