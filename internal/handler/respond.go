package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

func respondJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}

func callerID(r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.Header.Get("X-User-ID"))
	return id, err == nil
}

func parseLimit(r *http.Request, defaultLimit int) int {
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultLimit
}
