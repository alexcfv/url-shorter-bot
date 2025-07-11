package bot

import "sync"

type StateStore struct {
	mu     sync.RWMutex
	states map[int64]string
}

func NewStateStore() *StateStore {
	return &StateStore{states: make(map[int64]string)}
}

func (s *StateStore) Set(chatID int64, state string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[chatID] = state
}

func (s *StateStore) Get(chatID int64) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.states[chatID]
}

func (s *StateStore) Clear(chatID int64) {
	s.Set(chatID, "")
}
