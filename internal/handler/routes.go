package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/snaply/relation-service/internal/service"
	"go.uber.org/zap"
)

func NewRouter(follows service.FollowService, log *zap.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	followH := NewFollowHandler(follows, log)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1/relations", func(r chi.Router) {
		r.Route("/{user_id}", func(r chi.Router) {
			r.Post("/follow", followH.Follow)
			r.Delete("/follow", followH.Unfollow)
			r.Get("/status", followH.Status)
			r.Get("/counts", followH.Counts)
			r.Get("/followers", followH.Followers)
			r.Get("/following", followH.Following)
		})
	})

	return r
}
