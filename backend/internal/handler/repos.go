package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	ghclient "github.com/unified-dashboard/backend/internal/github"
)

type repositoriesResponse struct {
	Organization string                `json:"organization"`
	Count        int                   `json:"count"`
	Repositories []ghclient.Repository `json:"repositories"`
}

// ListRepositories returns repositories for the configured default organization.
func (h *Handler) ListRepositories(w http.ResponseWriter, r *http.Request) {
	h.listOrgRepositories(w, r, h.org)
}

// ListOrgRepositories returns repositories for :org path parameter.
func (h *Handler) ListOrgRepositories(w http.ResponseWriter, r *http.Request) {
	org := chi.URLParam(r, "org")
	if org == "" {
		writeError(w, http.StatusBadRequest, "organization is required")
		return
	}
	h.listOrgRepositories(w, r, org)
}

func (h *Handler) listOrgRepositories(w http.ResponseWriter, r *http.Request, org string) {
	key := "repos:" + org

	resp, hit, err := getOrFetch(h, r, key, func() (repositoriesResponse, error) {
		ctx, cancel := contextWithTimeout(r, 45*time.Second)
		defer cancel()

		repos, err := h.github.ListOrgRepositories(ctx, org)
		if err != nil {
			return repositoriesResponse{}, err
		}
		if repos == nil {
			repos = []ghclient.Repository{}
		}
		return repositoriesResponse{
			Organization: org,
			Count:        len(repos),
			Repositories: repos,
		}, nil
	})
	if err != nil {
		h.log.Error("failed to list repositories", "org", org, "error", err)
		writeError(w, http.StatusBadGateway, "failed to fetch repositories from GitHub")
		return
	}

	h.writeCached(w, r, key, hit, resp)
}
