package handler

import (
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/chi/v5"
	ghclient "github.com/unified-dashboard/backend/internal/github"
)

type trackingSummaryResponse struct {
	Organization string                   `json:"organization"`
	Summary      ghclient.TrackingMetrics `json:"summary"`
	ByRepository []ghclient.RepoMetricPoint `json:"by_repository"`
}

func (h *Handler) TrackingSummary(w http.ResponseWriter, r *http.Request) {
	key := "tracking-summary:" + h.org

	resp, hit, err := getOrFetch(h, r, key, func() (trackingSummaryResponse, error) {
		ctx, cancel := contextWithTimeout(r, 120*time.Second)
		defer cancel()

		summary, err := h.github.OrgTrackingSummary(ctx, h.org)
		if err != nil {
			return trackingSummaryResponse{}, err
		}

		repos, err := h.github.ListOrgRepositories(ctx, h.org)
		if err != nil {
			return trackingSummaryResponse{}, err
		}

		sort.Slice(repos, func(i, j int) bool {
			return repos[i].Stars > repos[j].Stars
		})
		const maxBreakdown = 5
		if len(repos) > maxBreakdown {
			repos = repos[:maxBreakdown]
		}

		breakdown, err := h.github.RepoTrackingBreakdown(ctx, repos)
		if err != nil && len(breakdown) == 0 {
			h.log.Error("failed repo tracking breakdown", "error", err)
		}
		if breakdown == nil {
			breakdown = []ghclient.RepoMetricPoint{}
		}
		if len(breakdown) > 0 {
			var failSum, minSum float64
			for _, p := range breakdown {
				failSum += p.Metrics.CIFailureRate
				minSum += p.Metrics.AvgCIMinutes
				summary.LinkedBugs += p.Metrics.LinkedBugs
				summary.DependencyAlerts += p.Metrics.DependencyAlerts
			}
			n := float64(len(breakdown))
			summary.CIFailureRate = failSum / n
			summary.AvgCIMinutes = minSum / n
		}

		return trackingSummaryResponse{
			Organization: h.org,
			Summary:      summary,
			ByRepository: breakdown,
		}, nil
	})
	if err != nil {
		h.log.Error("failed org tracking summary", "org", h.org, "error", err)
		writeError(w, http.StatusBadGateway, "failed to fetch tracking summary from GitHub")
		return
	}

	h.writeCached(w, r, key, hit, resp)
}

func (h *Handler) RepoTracking(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	name := chi.URLParam(r, "name")
	if owner == "" || name == "" {
		writeError(w, http.StatusBadRequest, "owner and repository name are required")
		return
	}

	key := "tracking:" + owner + "/" + name

	body, hit, err := getOrFetch(h, r, key, func() (map[string]any, error) {
		ctx, cancel := contextWithTimeout(r, 60*time.Second)
		defer cancel()

		metrics, err := h.github.RepoTracking(ctx, owner, name)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"repository": owner + "/" + name,
			"metrics":    metrics,
		}, nil
	})
	if err != nil {
		h.log.Error("failed repo tracking", "owner", owner, "repo", name, "error", err)
		writeError(w, http.StatusBadGateway, "failed to fetch repository tracking from GitHub")
		return
	}

	h.writeCached(w, r, key, hit, body)
}
