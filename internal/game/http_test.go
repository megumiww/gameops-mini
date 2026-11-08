package game

import (
	"encoding/json"
	"github.com/megumiww/gameops-mini/internal/store"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func testHandler(t *testing.T) Handler {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "games.json"))
	if err != nil {
		t.Fatal(err)
	}
	return Handler{Store: db}
}

func TestCreateAndListGames(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/games", strings.NewReader(`{"name":"Skyforge","region":"US"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	list := httptest.NewRequest(http.MethodGet, "/api/games", nil)
	listRec := httptest.NewRecorder()
	h.ServeHTTP(listRec, list)
	var games []store.Game
	if err := json.NewDecoder(listRec.Body).Decode(&games); err != nil {
		t.Fatal(err)
	}
	if len(games) != 1 || games[0].Name != "Skyforge" {
		t.Fatalf("unexpected games: %#v", games)
	}
}

func TestCreateGameValidatesRequiredFields(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/games", strings.NewReader(`{"name":""}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
