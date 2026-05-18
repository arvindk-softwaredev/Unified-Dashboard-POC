package handler

import (
	"net/http"

	"github.com/unified-dashboard/backend/internal/cache"
)

func (h *Handler) cacheBypass(r *http.Request) bool {
	return r.URL.Query().Get("refresh") == "true"
}

func (h *Handler) writeCached(w http.ResponseWriter, r *http.Request, key string, hit bool, body any) {
	if hit {
		w.Header().Set("X-Cache", "HIT")
	} else {
		w.Header().Set("X-Cache", "MISS")
	}
	writeJSON(w, http.StatusOK, body)
}

func getOrFetch[T any](h *Handler, r *http.Request, key string, fetch func() (T, error)) (T, bool, error) {
	if h.cacheBypass(r) {
		h.cache.Delete(key)
	}
	return cacheGetOrFetch(h.cache, key, fetch)
}

func cacheGetOrFetch[T any](c *cache.Cache, key string, fetch func() (T, error)) (T, bool, error) {
	return cache.GetOrFetch(c, key, fetch)
}
