package main

import (
	"fmt"
	"sync"
)

type ShortURL string
type LongURL string

type URLStore struct {
	urls map[ShortURL]LongURL
	mu   sync.RWMutex
}

func (s *URLStore) Get(key ShortURL) LongURL {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url := s.urls[key]
	return url
}

func (s *URLStore) Set(key, url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, present := s.urls[key]
	if present {
		return false
	}
	s.urls[key] = url
	return true
}
