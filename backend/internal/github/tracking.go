package github

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	gh "github.com/google/go-github/v69/github"
)

type TrackingMetrics struct {
	GoodFirstIssues  int     `json:"good_first_issues"`
	OpenBugs         int     `json:"open_bugs"`
	LinkedBugs       int     `json:"linked_bugs"`
	PendingPRs       int     `json:"pending_prs"`
	DependencyAlerts int     `json:"dependency_alerts"`
	CIFailureRate    float64 `json:"ci_failure_rate"`
	AvgCIMinutes     float64 `json:"avg_ci_minutes"`
	WorkflowRuns     int     `json:"workflow_runs_sampled"`
}

var bugLabels = []string{"bug", "kind/bug", "kind/bugs"}
var issueRefPattern = regexp.MustCompile(`(?i)(?:fixes|closes|refs|related to)?\s*#(\d+)`)

func (c *Client) OrgTrackingSummary(ctx context.Context, org string) (TrackingMetrics, error) {
	// Search API is limited to 30 req/min (auth) — pace calls to avoid secondary rate limits.
	queries := []string{
		fmt.Sprintf("org:%s is:issue is:open label:\"good first issue\"", org),
		fmt.Sprintf("org:%s is:issue is:open label:bug", org),
		fmt.Sprintf("org:%s is:pr is:open", org),
	}
	counts := make([]int, len(queries))
	for i, q := range queries {
		if err := sleepCtx(ctx, time.Second); err != nil {
			return TrackingMetrics{}, err
		}
		n, err := c.searchIssueCount(ctx, q)
		if err != nil {
			return TrackingMetrics{}, err
		}
		counts[i] = n
	}

	return TrackingMetrics{
		GoodFirstIssues: counts[0],
		OpenBugs:        counts[1],
		PendingPRs:      counts[2],
	}, nil
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (c *Client) RepoTracking(ctx context.Context, owner, repo string) (TrackingMetrics, error) {
	metrics := TrackingMetrics{}

	gfi, err := c.countIssuesByLabels(ctx, owner, repo, goodFirstLabels)
	if err != nil {
		return metrics, err
	}
	metrics.GoodFirstIssues = gfi

	bugs, linked, err := c.countBugs(ctx, owner, repo)
	if err != nil {
		return metrics, err
	}
	metrics.OpenBugs = bugs
	metrics.LinkedBugs = linked

	prs, err := c.countOpenPRs(ctx, owner, repo)
	if err != nil {
		return metrics, err
	}
	metrics.PendingPRs = prs

	alerts, err := c.countDependabotAlerts(ctx, owner, repo)
	if err != nil {
		alerts = 0
	}
	metrics.DependencyAlerts = alerts

	ciRate, avgMin, sampled, err := c.workflowMetrics(ctx, owner, repo)
	if err != nil {
		ciRate, avgMin, sampled = 0, 0, 0
	}
	metrics.CIFailureRate = ciRate
	metrics.AvgCIMinutes = avgMin
	metrics.WorkflowRuns = sampled

	return metrics, nil
}

func (c *Client) searchIssueCount(ctx context.Context, query string) (int, error) {
	result, _, err := c.api.Search.Issues(ctx, query, &gh.SearchOptions{ListOptions: gh.ListOptions{PerPage: 1}})
	if err != nil {
		return 0, fmt.Errorf("search: %w", err)
	}
	return result.GetTotal(), nil
}

func (c *Client) countIssuesByLabels(ctx context.Context, owner, repo string, labels []string) (int, error) {
	seen := make(map[int64]struct{})
	total := 0
	for _, label := range labels {
		opts := &gh.IssueListByRepoOptions{
			State:       "open",
			Labels:      []string{label},
			ListOptions: gh.ListOptions{PerPage: 100},
		}
		for {
			issues, resp, err := c.api.Issues.ListByRepo(ctx, owner, repo, opts)
			if err != nil {
				return 0, err
			}
			for _, issue := range issues {
				if issue.PullRequestLinks != nil {
					continue
				}
				if _, ok := seen[issue.GetID()]; ok {
					continue
				}
				seen[issue.GetID()] = struct{}{}
				total++
			}
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
	}
	return total, nil
}

func (c *Client) countBugs(ctx context.Context, owner, repo string) (open, linked int, err error) {
	seen := make(map[int64]*gh.Issue)
	for _, label := range bugLabels {
		opts := &gh.IssueListByRepoOptions{
			State:       "open",
			Labels:      []string{label},
			ListOptions: gh.ListOptions{PerPage: 100},
		}
		for {
			issues, resp, err := c.api.Issues.ListByRepo(ctx, owner, repo, opts)
			if err != nil {
				return 0, 0, err
			}
			for _, issue := range issues {
				if issue.PullRequestLinks != nil {
					continue
				}
				seen[issue.GetID()] = issue
			}
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
	}
	for _, issue := range seen {
		open++
		body := issue.GetBody()
		if body != "" && issueRefPattern.MatchString(body) {
			linked++
		}
	}
	return open, linked, nil
}

func (c *Client) countOpenPRs(ctx context.Context, owner, repo string) (int, error) {
	opts := &gh.PullRequestListOptions{
		State:       "open",
		ListOptions: gh.ListOptions{PerPage: 100},
	}
	total := 0
	for {
		prs, resp, err := c.api.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return 0, err
		}
		total += len(prs)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return total, nil
}

func (c *Client) countDependabotAlerts(ctx context.Context, owner, repo string) (int, error) {
	opts := &gh.ListAlertsOptions{ListOptions: gh.ListOptions{PerPage: 100}}
	total := 0
	for {
		alerts, resp, err := c.api.Dependabot.ListRepoAlerts(ctx, owner, repo, opts)
		if err != nil {
			return 0, err
		}
		for _, a := range alerts {
			if strings.EqualFold(a.GetState(), "open") {
				total++
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}
	return total, nil
}

func (c *Client) workflowMetrics(ctx context.Context, owner, repo string) (failureRate float64, avgMinutes float64, sampled int, err error) {
	opts := &gh.ListWorkflowRunsOptions{
		ListOptions: gh.ListOptions{PerPage: 30},
	}
	runs, _, err := c.api.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, opts)
	if err != nil {
		return 0, 0, 0, err
	}

	var failures int
	var totalMinutes float64
	var measured int

	for _, run := range runs.WorkflowRuns {
		sampled++
		switch strings.ToLower(run.GetConclusion()) {
		case "failure", "timed_out", "cancelled":
			failures++
		}
		if run.RunStartedAt != nil && run.UpdatedAt != nil {
			d := run.UpdatedAt.Sub(run.RunStartedAt.Time)
			if d > 0 {
				totalMinutes += d.Minutes()
				measured++
			}
		}
	}

	if sampled > 0 {
		failureRate = float64(failures) / float64(sampled) * 100
	}
	if measured > 0 {
		avgMinutes = totalMinutes / float64(measured)
	}
	return failureRate, avgMinutes, sampled, nil
}

// RepoTrackingBreakdown returns per-repo metrics for charting across the org.
func (c *Client) RepoTrackingBreakdown(ctx context.Context, repos []Repository) ([]RepoMetricPoint, error) {
	points := make([]RepoMetricPoint, 0, len(repos))
	for _, repo := range repos {
		parts := strings.SplitN(repo.FullName, "/", 2)
		if len(parts) != 2 {
			continue
		}
		m, err := c.RepoTracking(ctx, parts[0], parts[1])
		if err != nil {
			continue
		}
		points = append(points, RepoMetricPoint{
			Repository: repo.Name,
			Metrics:    m,
		})
		// Respect rate limits for unauthenticated API
		select {
		case <-ctx.Done():
			return points, ctx.Err()
		case <-time.After(300 * time.Millisecond):
		}
	}
	return points, nil
}

type RepoMetricPoint struct {
	Repository string          `json:"repository"`
	Metrics    TrackingMetrics `json:"metrics"`
}
