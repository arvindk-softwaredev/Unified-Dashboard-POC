package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	ghclient "github.com/unified-dashboard/backend/internal/github"
	"github.com/unified-dashboard/backend/internal/llm"
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

	aiMode := r.URL.Query().Get("ai_mode") == "true"

	key := "good-first-issues:" + owner + "/" + name
	if aiMode {
		key = "good-first-issues:ai:" + owner + "/" + name
	}

	resp, hit, err := getOrFetch(h, r, key, func() (goodFirstIssuesResponse, error) {
		ctx, cancel := contextWithTimeout(r, 60*time.Second)
		defer cancel()

		issues, err := h.github.ListGoodFirstIssues(ctx, owner, name)
		if err != nil {
			return goodFirstIssuesResponse{}, err
		}
		if issues == nil {
			issues = []ghclient.Issue{}
		}

		if aiMode && h.llm.Enabled() && len(issues) > 0 {
			repoDesc, _ := h.github.GetRepoDescription(ctx, owner, name)
			readme, _ := h.github.GetReadmeContent(ctx, owner, name)

			inputs := make([]llm.IssueInput, len(issues))
			for i, issue := range issues {
				inputs[i] = llm.IssueInput{
					Number: issue.Number,
					Title:  issue.Title,
					Body:   issue.Body,
				}
			}

			complexities, err := h.llm.AnalyzeIssueComplexity(ctx, repoDesc, readme, inputs)
			if err != nil {
				h.log.Error("LLM complexity analysis failed for good-first-issues", "owner", owner, "repo", name, "error", err)
			} else {
				for i, issue := range issues {
					if c, ok := complexities[issue.Number]; ok {
						issues[i].Complexity = c
					}
				}
			}
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
