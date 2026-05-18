package github

import (
	"context"
	"fmt"
	"sort"
	"time"

	gh "github.com/google/go-github/v69/github"
)

type MonthlyBucket struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type PullRequestItem struct {
	ID        int64  `json:"id"`
	Number    int    `json:"number"`
	Title     string `json:"title"`
	HTMLURL   string `json:"html_url"`
	CreatedAt string `json:"created_at,omitempty"`
}

type CategoryInsight struct {
	Key          string            `json:"key"`
	Label        string            `json:"label"`
	Total        int               `json:"total"`
	Monthly      []MonthlyBucket   `json:"monthly"`
	Issues       []Issue           `json:"issues,omitempty"`
	PullRequests []PullRequestItem `json:"pull_requests,omitempty"`
}

type RepoInsights struct {
	Repository string            `json:"repository"`
	Metrics    TrackingMetrics   `json:"metrics"`
	Categories []CategoryInsight `json:"categories"`
}

func (c *Client) RepoInsights(ctx context.Context, owner, repo string) (*RepoInsights, error) {
	gfi, err := c.ListGoodFirstIssues(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	bugs, err := c.listBugIssues(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	linked := filterLinkedIssues(bugs)

	prs, err := c.listOpenPullRequests(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	deps, _ := c.countDependabotAlerts(ctx, owner, repo)
	ciRate, avgMin, runs, _ := c.workflowMetrics(ctx, owner, repo)

	categories := []CategoryInsight{
		{
			Key:     "good_first_issues",
			Label:   "Good first issues",
			Total:   len(gfi),
			Monthly: groupIssuesByMonth(gfi),
			Issues:  gfi,
		},
		{
			Key:     "open_bugs",
			Label:   "Open bugs",
			Total:   len(bugs),
			Monthly: groupIssuesByMonth(bugs),
			Issues:  bugs,
		},
		{
			Key:     "linked_bugs",
			Label:   "Linked bugs",
			Total:   len(linked),
			Monthly: groupIssuesByMonth(linked),
			Issues:  linked,
		},
		{
			Key:          "pending_prs",
			Label:        "Pending PRs",
			Total:        len(prs),
			Monthly:      groupPRsByMonth(prs),
			PullRequests: prs,
		},
		{
			Key:     "dependency_alerts",
			Label:   "Dependency alerts",
			Total:   deps,
			Monthly: singleMonthBucket(deps),
		},
	}

	return &RepoInsights{
		Repository: owner + "/" + repo,
		Metrics: TrackingMetrics{
			GoodFirstIssues:     len(gfi),
			OpenBugs:            len(bugs),
			LinkedBugs:          len(linked),
			PendingPRs:          len(prs),
			DependencyAlerts:    deps,
			CIFailureRate:       ciRate,
			AvgCIMinutes:        avgMin,
			WorkflowRuns:        runs,
		},
		Categories: categories,
	}, nil
}

func (c *Client) listBugIssues(ctx context.Context, owner, repo string) ([]Issue, error) {
	seen := make(map[int64]Issue)
	for _, label := range bugLabels {
		issues, err := c.listIssuesByLabel(ctx, owner, repo, label)
		if err != nil {
			return nil, err
		}
		for _, issue := range issues {
			seen[issue.ID] = issue
		}
	}
	out := make([]Issue, 0, len(seen))
	for _, issue := range seen {
		out = append(out, issue)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Number > out[j].Number })
	return out, nil
}

func (c *Client) listOpenPullRequests(ctx context.Context, owner, repo string) ([]PullRequestItem, error) {
	opts := &gh.PullRequestListOptions{
		State:       "open",
		ListOptions: gh.ListOptions{PerPage: 100},
	}
	var out []PullRequestItem
	for {
		prs, resp, err := c.api.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("list pull requests: %w", err)
		}
		for _, pr := range prs {
			item := PullRequestItem{
				ID:      pr.GetID(),
				Number:  pr.GetNumber(),
				Title:   pr.GetTitle(),
				HTMLURL: pr.GetHTMLURL(),
			}
			if pr.CreatedAt != nil {
				item.CreatedAt = pr.CreatedAt.Format(time.RFC3339)
			}
			out = append(out, item)
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return out, nil
}

func filterLinkedIssues(issues []Issue) []Issue {
	out := make([]Issue, 0)
	for _, issue := range issues {
		if issue.Body != "" && issueRefPattern.MatchString(issue.Body) {
			out = append(out, issue)
		}
	}
	return out
}

func groupIssuesByMonth(issues []Issue) []MonthlyBucket {
	return groupByMonth(func() []time.Time {
		var dates []time.Time
		for _, i := range issues {
			if t, ok := parseMonth(i.CreatedAt); ok {
				dates = append(dates, t)
			}
		}
		return dates
	}())
}

func groupPRsByMonth(prs []PullRequestItem) []MonthlyBucket {
	var dates []time.Time
	for _, pr := range prs {
		if t, ok := parseMonth(pr.CreatedAt); ok {
			dates = append(dates, t)
		}
	}
	return groupByMonth(dates)
}

func singleMonthBucket(count int) []MonthlyBucket {
	if count == 0 {
		return []MonthlyBucket{}
	}
	return []MonthlyBucket{{Month: time.Now().Format("2006-01"), Count: count}}
}

func parseMonth(iso string) (time.Time, bool) {
	if iso == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return time.Time{}, false
	}
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), true
}

func groupByMonth(dates []time.Time) []MonthlyBucket {
	counts := map[string]int{}
	for _, d := range dates {
		key := d.Format("2006-01")
		counts[key]++
	}
	months := make([]string, 0, len(counts))
	for m := range counts {
		months = append(months, m)
	}
	sort.Strings(months)
	out := make([]MonthlyBucket, len(months))
	for i, m := range months {
		out[i] = MonthlyBucket{Month: m, Count: counts[m]}
	}
	return out
}
