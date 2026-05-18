package handler

import "net/http"

func (h *Handler) Hello(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]string{
			"message": "Hello from Unified Dashboard API",
			"method":  "GET",
		})
	case http.MethodPut:
		writeJSON(w, http.StatusOK, map[string]string{
			"message": "Hello from Unified Dashboard API",
			"method":  "PUT",
		})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	entries, _ := h.cache.Stats()
	writeJSON(w, http.StatusOK, map[string]any{
		"status":                 "ok",
		"github_authenticated":   h.github.Authenticated(),
		"cache_ttl_minutes":      h.cache.TTL().Minutes(),
		"cache_entries":          entries,
	})
}
