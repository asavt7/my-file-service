package server

import (
	"context"
	"net/http"
	"time"
)

func (s *APIServer) timeoutMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 10*time.Second)
		defer cancel()
		req.WithContext(ctx)
		h.ServeHTTP(rw, req)
	})
}
