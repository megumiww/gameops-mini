package game

import (
	"encoding/json"
	"errors"
	"github.com/megumiww/gameops-mini/internal/store"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct{ Store *store.Store }
type createGameRequest struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/games" {
		switch r.Method {
		case http.MethodGet:
			h.list(w)
		case http.MethodPost:
			h.create(w, r)
		default:
			methodNotAllowed(w, http.MethodGet, http.MethodPost)
		}
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/games/") && r.Method == http.MethodGet {
		h.get(w, r)
		return
	}
	http.NotFound(w, r)
}

func (h Handler) list(w http.ResponseWriter) { writeJSON(w, http.StatusOK, h.Store.ListGames()) }

func (h Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/games/"))
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "game id must be a positive integer")
		return
	}
	game, err := h.Store.GetGame(id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "game not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not read game")
		return
	}
	writeJSON(w, http.StatusOK, game)
}

func (h Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "request body must be valid JSON")
		return
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Region) == "" {
		writeError(w, http.StatusBadRequest, "name and region are required")
		return
	}
	game, err := h.Store.CreateGame(strings.TrimSpace(req.Name), strings.TrimSpace(req.Region))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not save game")
		return
	}
	writeJSON(w, http.StatusCreated, game)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
func methodNotAllowed(w http.ResponseWriter, methods ...string) {
	w.Header().Set("Allow", strings.Join(methods, ", "))
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}
