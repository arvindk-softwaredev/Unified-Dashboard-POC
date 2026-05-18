package handler

import (
	"context"
	"net/http"
	"time"
)

func contextWithTimeout(r *http.Request, d time.Duration) (context.Context, context.CancelFunc) {
	if deadline, ok := r.Context().Deadline(); ok && time.Until(deadline) < d {
		return context.WithTimeout(r.Context(), time.Until(deadline))
	}
	return context.WithTimeout(r.Context(), d)
}
