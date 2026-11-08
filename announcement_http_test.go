package announcement

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnnouncementAdminFlow(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "announcements.json"))
	if err != nil {
		t.Fatal(err)
	}
	h := Handler{Store: db, AdminToken: "test-token"}
	unauthorized := httptest.NewRequest(http.MethodPost, "/api/admin/announcements", strings.NewReader(`{"game_id":1}`))
	unauthRec := httptest.NewRecorder()
	h.ServeHTTP(unauthRec, unauthorized)
	if unauthRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", unauthRec.Code)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/admin/announcements", strings.NewReader(`{"game_id":1,"title":"Launch","content":"Welcome","language":"en","region":"US"}`))
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	list := httptest.NewRequest(http.MethodGet, "/api/games/1/announcements", nil)
	listRec := httptest.NewRecorder()
	h.ServeHTTP(listRec, list)
	if listRec.Code != http.StatusOK || !strings.Contains(listRec.Body.String(), "Launch") || !strings.Contains(listRec.Body.String(), `"total":1`) {
		t.Fatalf("unexpected list response: %s", listRec.Body.String())
	}
}
