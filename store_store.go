package store

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

var ErrNotFound = errors.New("game not found")

type Game struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Region string `json:"region"`
}

type Store struct {
	mu     sync.RWMutex
	path   string
	nextID int
	games  []Game
}

func Open(path string) (*Store, error) {
	s := &Store{path: path, nextID: 1, games: []Game{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(data, &s.games); err != nil {
		return nil, err
	}
	for _, game := range s.games {
		if game.ID >= s.nextID {
			s.nextID = game.ID + 1
		}
	}
	return s, nil
}

func (s *Store) ListGames() []Game {
	s.mu.RLock()
	defer s.mu.RUnlock()
	games := make([]Game, len(s.games))
	copy(games, s.games)
	return games
}

func (s *Store) GetGame(id int) (Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, game := range s.games {
		if game.ID == id {
			return game, nil
		}
	}
	return Game{}, ErrNotFound
}

func (s *Store) CreateGame(name, region string) (Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	game := Game{ID: s.nextID, Name: name, Region: region}
	s.nextID++
	s.games = append(s.games, game)
	data, err := json.MarshalIndent(s.games, "", "  ")
	if err != nil {
		return Game{}, err
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return Game{}, err
	}
	return game, nil
}
