package announcement

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	Store      *Store
	AdminToken string
}
type createRequest struct {
	GameID   int    `json:"game_id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Language string `json:"language"`
	Region   string `json:"region"`
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/games/") && strings.HasSuffix(r.URL.Path, "/announcements") && r.Method == http.MethodGet {
		h.listByGame(w, r)
		return
	}
	if r.URL.Path == "/api/admin/announcements" && r.Method == http.MethodPost {
		h.create(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/admin/announcements/") && r.Method == http.MethodDelete {
		h.delete(w, r)
		return
	}
	http.NotFound(w, r)
}

func (h Handler) listByGame(w http.ResponseWriter, r *http.Request) {
	part := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/games/"), "/announcements")
	id, err := strconv.Atoi(part)
	if err != nil || id < 1 {
		writeError(w, 400, "game id must be a positive integer")
		return
	}
	page, pageSize := 1, 10
	if value, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && value > 0 {
		page = value
	}
	if value, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil && value > 0 && value <= 100 {
		pageSize = value
	}
	items, total := h.Store.ListByGame(id, r.URL.Query().Get("region"), r.URL.Query().Get("language"), (page-1)*pageSize, pageSize)
	writeJSON(w, 200, map[string]any{"items": items, "page": page, "page_size": pageSize, "total": total})
}

func (h Handler) create(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		writeError(w, http.StatusUnauthorized, "admin token required")
		return
	}
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "request body must be valid JSON")
		return
	}
	if req.GameID < 1 || strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Content) == "" || strings.TrimSpace(req.Language) == "" || strings.TrimSpace(req.Region) == "" {
		writeError(w, 400, "game_id, title, content, language and region are required")
		return
	}
	item, err := h.Store.Create(req.GameID, strings.TrimSpace(req.Title), strings.TrimSpace(req.Content), strings.TrimSpace(req.Language), strings.TrimSpace(req.Region))
	if err != nil {
		writeError(w, 500, "could not save announcement")
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h Handler) delete(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		writeError(w, http.StatusUnauthorized, "admin token required")
		return
	}
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/admin/announcements/"))
	if err != nil || id < 1 {
		writeError(w, 400, "announcement id must be a positive integer")
		return
	}
	if err := h.Store.Delete(id); errors.Is(err, ErrNotFound) {
		writeError(w, 404, "announcement not found")
		return
	} else if err != nil {
		writeError(w, 500, "could not delete announcement")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) authorized(r *http.Request) bool {
	return h.AdminToken != "" && r.Header.Get("Authorization") == "Bearer "+h.AdminToken
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
