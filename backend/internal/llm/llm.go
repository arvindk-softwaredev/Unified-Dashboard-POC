package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

const chunkSize = 25

type Client struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     strings.TrimSpace(apiKey),
		httpClient: &http.Client{Timeout: 120 * time.Second},
		model:      "gemini-2.0-flash",
	}
}

func (c *Client) Enabled() bool {
	return c.apiKey != ""
}

type IssueInput struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type issueComplexityResult struct {
	Number     int    `json:"number"`
	Complexity string `json:"complexity"`
}

// AnalyzeIssueComplexity splits issues into chunks and processes them
// concurrently via Gemini, merging all results into a single map.
func (c *Client) AnalyzeIssueComplexity(ctx context.Context, repoDescription, readmeContent string, issues []IssueInput) (map[int]string, error) {
	if !c.Enabled() || len(issues) == 0 {
		return map[int]string{}, nil
	}

	readme := readmeContent
	if len(readme) > 4000 {
		readme = readme[:4000] + "\n... (truncated)"
	}

	trimmed := make([]IssueInput, len(issues))
	for i, issue := range issues {
		body := issue.Body
		if len(body) > 500 {
			body = body[:500] + "..."
		}
		trimmed[i] = IssueInput{Number: issue.Number, Title: issue.Title, Body: body}
	}

	chunks := chunkIssues(trimmed, chunkSize)

	slog.Info("LLM complexity analysis starting",
		"total_issues", len(trimmed),
		"chunks", len(chunks),
		"chunk_size", chunkSize,
	)

	var (
		mu     sync.Mutex
		merged = make(map[int]string, len(trimmed))
		wg     sync.WaitGroup
		errors []error
	)

	for i, chunk := range chunks {
		wg.Add(1)
		go func(idx int, batch []IssueInput) {
			defer wg.Done()

			result, err := c.analyzeChunk(ctx, repoDescription, readme, batch)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				slog.Warn("LLM chunk failed", "chunk", idx, "size", len(batch), "error", err)
				errors = append(errors, err)
				return
			}
			for k, v := range result {
				merged[k] = v
			}
			slog.Info("LLM chunk completed", "chunk", idx, "classified", len(result))
		}(i, chunk)
	}

	wg.Wait()

	if len(merged) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("all %d LLM chunks failed; first error: %w", len(errors), errors[0])
	}

	slog.Info("LLM complexity analysis done", "classified", len(merged), "failed_chunks", len(errors))
	return merged, nil
}

func chunkIssues(issues []IssueInput, size int) [][]IssueInput {
	var chunks [][]IssueInput
	for i := 0; i < len(issues); i += size {
		end := i + size
		if end > len(issues) {
			end = len(issues)
		}
		chunks = append(chunks, issues[i:end])
	}
	return chunks
}

// analyzeChunk sends a single batch of issues to Gemini and parses the response.
func (c *Client) analyzeChunk(ctx context.Context, repoDescription, readme string, issues []IssueInput) (map[int]string, error) {
	issuesJSON, _ := json.Marshal(issues)

	prompt := fmt.Sprintf(`You are an expert software engineer. Analyze GitHub issues and classify their complexity level.
For each issue, determine the difficulty level based on the issue title, description, and the repository context.

Classify each issue as exactly one of:
- "beginner" — Simple changes like documentation fixes, typos, config changes, adding tests for existing code, small UI tweaks
- "intermediate" — Moderate changes requiring understanding of the codebase, feature enhancements, bug fixes with some debugging needed
- "advanced" — Complex changes involving architecture, performance optimization, security fixes, cross-cutting concerns, or deep domain knowledge

Respond with a JSON array of objects with "number" (issue number) and "complexity" (one of: beginner, intermediate, advanced).
Respond ONLY with the JSON array, no markdown fences, no explanation.

Repository description: %s

Repository documentation (README):
%s

Issues to analyze:
%s`, repoDescription, readme, string(issuesJSON))

	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		c.model, c.apiKey,
	)

	reqBody, _ := json.Marshal(geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:      0.1,
			MaxOutputTokens:  4096,
			ResponseMimeType: "application/json",
		},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API error %d: %s", resp.StatusCode, string(respBody))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no candidates in gemini response")
	}

	content := strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var results []issueComplexityResult
	if err := json.Unmarshal([]byte(content), &results); err != nil {
		slog.Warn("failed to parse LLM chunk response", "error", err, "content_len", len(content))
		return nil, fmt.Errorf("parse complexity results: %w", err)
	}

	out := make(map[int]string, len(results))
	for _, r := range results {
		switch r.Complexity {
		case "beginner", "intermediate", "advanced":
			out[r.Number] = r.Complexity
		default:
			slog.Warn("unexpected complexity value from LLM", "number", r.Number, "complexity", r.Complexity)
		}
	}
	return out, nil
}

// Gemini API request/response types

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature      float64 `json:"temperature"`
	MaxOutputTokens  int     `json:"maxOutputTokens"`
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}
