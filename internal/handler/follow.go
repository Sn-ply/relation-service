package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/snaply/relation-service/internal/service"
	"go.uber.org/zap"
)

type FollowHandler struct {
	follows service.FollowService
	log     *zap.Logger
}

func NewFollowHandler(follows service.FollowService, log *zap.Logger) *FollowHandler {
	return &FollowHandler{follows: follows, log: log}
}

func targetUserID(r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "user_id"))
	return id, err == nil
}

func (h *FollowHandler) Follow(w http.ResponseWriter, r *http.Request) {
	callerID, ok := callerID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "missing user identity")
		return
	}
	targetID, ok := targetUserID(r)
	if !ok {
		respondError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	if err := h.follows.Follow(r.Context(), callerID, targetID); err != nil {
		if errors.Is(err, service.ErrCannotFollowSelf) {
			respondError(w, http.StatusBadRequest, "cannot follow yourself")
			return
		}
		h.log.Error("follow error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FollowHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	callerID, ok := callerID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "missing user identity")
		return
	}
	targetID, ok := targetUserID(r)
	if !ok {
		respondError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	if err := h.follows.Unfollow(r.Context(), callerID, targetID); err != nil {
		h.log.Error("unfollow error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FollowHandler) Status(w http.ResponseWriter, r *http.Request) {
	callerID, ok := callerID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "missing user identity")
		return
	}
	targetID, ok := targetUserID(r)
	if !ok {
		respondError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	following, err := h.follows.Status(r.Context(), callerID, targetID)
	if err != nil {
		h.log.Error("status error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"following": following})
}

func (h *FollowHandler) Counts(w http.ResponseWriter, r *http.Request) {
	targetID, ok := targetUserID(r)
	if !ok {
		respondError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	counts, err := h.follows.Counts(r.Context(), targetID)
	if err != nil {
		h.log.Error("counts error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondJSON(w, http.StatusOK, counts)
}

func (h *FollowHandler) Followers(w http.ResponseWriter, r *http.Request) {
	targetID, ok := targetUserID(r)
	if !ok {
		respondError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	page, err := h.follows.Followers(r.Context(), targetID, r.URL.Query().Get("cursor"), parseLimit(r, 20))
	if err != nil {
		h.log.Error("followers error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"data":        page.UserIDs,
		"next_cursor": page.NextCursor,
	})
}

func (h *FollowHandler) Following(w http.ResponseWriter, r *http.Request) {
	targetID, ok := targetUserID(r)
	if !ok {
		respondError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	page, err := h.follows.Following(r.Context(), targetID, r.URL.Query().Get("cursor"), parseLimit(r, 20))
	if err != nil {
		h.log.Error("following error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"data":        page.UserIDs,
		"next_cursor": page.NextCursor,
	})
}
