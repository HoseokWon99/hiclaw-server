package storage

import (
	"sync"

	"hiclaw-server/core"
)

type MemoryChatStore struct {
	mu       sync.Mutex
	messages []*core.Message
}

func NewMemoryChatStore() *MemoryChatStore {
	return &MemoryChatStore{}
}

func (s *MemoryChatStore) Save(msg *core.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, msg)
	return nil
}

func (s *MemoryChatStore) ListRecent(limit int) ([]*core.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	n := len(s.messages)
	if limit > n {
		limit = n
	}
	result := make([]*core.Message, limit)
	for i := 0; i < limit; i++ {
		result[i] = s.messages[n-1-i]
	}
	return result, nil
}
