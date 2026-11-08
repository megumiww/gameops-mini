package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/megumiww/gameops-mini/internal/announcement"
	"github.com/megumiww/gameops-mini/internal/game"
	"github.com/megumiww/gameops-mini/internal/store"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(healthResponse{
		Status:  "ok",
		Service: "gameops-mini",
	})
}

func newServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler)
	dataFile := os.Getenv("GAMEOPS_DATA_FILE")
	if dataFile == "" {
		dataFile = "gameops-data.json"
	}
	dataStore, err := store.Open(dataFile)
	if err != nil {
		log.Fatalf("open data store: %v", err)
	}
	gameHandler := game.Handler{Store: dataStore}
	mux.Handle("/api/games", gameHandler)
	announcementStore, err := announcement.Open(dataFile + ".announcements")
	if err != nil {
		log.Fatalf("open announcement store: %v", err)
	}
	adminToken := os.Getenv("GAMEOPS_ADMIN_TOKEN")
	if adminToken == "" {
		adminToken = "dev-token"
	}
	announcementHandler := announcement.Handler{Store: announcementStore, AdminToken: adminToken}
	mux.Handle("/api/games/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/announcements") {
			announcementHandler.ServeHTTP(w, r)
			return
		}
		gameHandler.ServeHTTP(w, r)
	}))
	mux.Handle("/api/admin/announcements", announcementHandler)
	mux.Handle("/api/admin/announcements/", announcementHandler)

	addr := os.Getenv("GAMEOPS_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

func main() {
	server := newServer()
	var requestCounter uint64
	server.Handler = requestIDMiddleware(loggingMiddleware(server.Handler, &requestCounter))

	log.Printf("gameops-mini listening on %s", server.Addr)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("gameops-mini stopped")
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = fmt.Sprintf("req-%d", time.Now().UnixNano())
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler, counter *uint64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		n := atomic.AddUint64(counter, 1)
		log.Printf("request=%s method=%s path=%s duration=%s total=%d", w.Header().Get("X-Request-ID"), r.Method, r.URL.Path, time.Since(start), n)
	})
}
