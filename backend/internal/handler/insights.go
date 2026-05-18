package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	ghclient "github.com/unified-dashboard/backend/internal/github"
	"github.com/unified-dashboard/backend/internal/llm"
)

func (h *Handler) RepoInsights(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	name := chi.URLParam(r, "name")
	if owner == "" || name == "" {
		writeError(w, http.StatusBadRequest, "owner and repository name are required")
		return
	}

	aiMode := r.URL.Query().Get("ai_mode") == "true"

	key := "insights:" + owner + "/" + name
	if aiMode {
		key = "insights:ai:" + owner + "/" + name
	}

	insights, hit, err := getOrFetch(h, r, key, func() (*ghclient.RepoInsights, error) {
		ctx, cancel := contextWithTimeout(r, 120*time.Second)
		defer cancel()

		result, err := h.github.RepoInsights(ctx, owner, name)
		if err != nil {
			return nil, err
		}

		if aiMode && h.llm.Enabled() {
			h.enrichInsightsWithComplexity(ctx, owner, name, result)
		}

		return result, nil
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

func (h *Handler) enrichInsightsWithComplexity(ctx context.Context, owner, name string, insights *ghclient.RepoInsights) {
	repoDesc, _ := h.github.GetRepoDescription(ctx, owner, name)
	readme, _ := h.github.GetReadmeContent(ctx, owner, name)

	var allIssues []llm.IssueInput
	for _, cat := range insights.Categories {
		for _, issue := range cat.Issues {
			allIssues = append(allIssues, llm.IssueInput{
				Number: issue.Number,
				Title:  issue.Title,
				Body:   issue.Body,
			})
		}
	}

	if len(allIssues) == 0 {
		return
	}

	complexities, err := h.llm.AnalyzeIssueComplexity(ctx, repoDesc, readme, allIssues)
	if err != nil {
		h.log.Error("LLM complexity analysis failed", "owner", owner, "repo", name, "error", err)
		return
	}

	for i, cat := range insights.Categories {
		for j, issue := range cat.Issues {
			if c, ok := complexities[issue.Number]; ok {
				insights.Categories[i].Issues[j].Complexity = c
			}
		}
	}
}
