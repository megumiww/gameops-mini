package announcement

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

var ErrNotFound = errors.New("announcement not found")

type Announcement struct {
	ID          int    `json:"id"`
	GameID      int    `json:"game_id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Language    string `json:"language"`
	Region      string `json:"region"`
	PublishedAt string `json:"published_at"`
}

type Store struct {
	mu     sync.RWMutex
	path   string
	nextID int
	items  []Announcement
}

func Open(path string) (*Store, error) {
	s := &Store{path: path, nextID: 1, items: []Announcement{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &s.items); err != nil {
			return nil, err
		}
	}
	for _, item := range s.items {
		if item.ID >= s.nextID {
			s.nextID = item.ID + 1
		}
	}
	return s, nil
}

func (s *Store) ListByGame(gameID int, region, language string, offset, limit int) ([]Announcement, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	filtered := make([]Announcement, 0)
	for _, item := range s.items {
		if item.GameID == gameID && (region == "" || item.Region == region) && (language == "" || item.Language == language) {
			filtered = append(filtered, item)
		}
	}
	total := len(filtered)
	if offset >= total {
		return []Announcement{}, total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return filtered[offset:end], total
}

func (s *Store) Create(gameID int, title, content, language, region string) (Announcement, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := Announcement{ID: s.nextID, GameID: gameID, Title: title, Content: content, Language: language, Region: region, PublishedAt: time.Now().UTC().Format(time.RFC3339)}
	s.nextID++
	s.items = append(s.items, item)
	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return Announcement{}, err
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return Announcement{}, err
	}
	return item, nil
}

func (s *Store) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, item := range s.items {
		if item.ID == id {
			s.items = append(s.items[:i], s.items[i+1:]...)
			data, err := json.MarshalIndent(s.items, "", "  ")
			if err != nil {
				return err
			}
			return os.WriteFile(s.path, data, 0644)
		}
	}
	return ErrNotFound
}
