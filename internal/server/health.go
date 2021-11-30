package server

import (
	"net/http"
)

// readiness
// @Summary readiness
// @Description readiness probe
// @Tags health
// @Produce text/html
// @Success 200
// @Router /api/health/readiness [get]
func (s *APIServer) readiness(w http.ResponseWriter, r *http.Request) {
	err := s.health.ReadinessProbe(r.Context())
	if err != nil {
		http.Error(w, "service not ready", 500)
		return
	}
	w.WriteHeader(200)
	return
}

// liveness
// @Summary liveness
// @Description liveness probe
// @Tags health
// @Produce text/html
// @Success 200
// @Router /api/health/liveness [get]
func (s *APIServer) liveness(w http.ResponseWriter, r *http.Request) {
	err := s.health.LivenessProbe(r.Context())
	if err != nil {
		http.Error(w, "service is unhealthy", 500)
		return
	}
	w.WriteHeader(200)
	return
}
