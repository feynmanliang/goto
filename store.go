package main

import (
	"crypto/md5"
	"fmt"
	"strings"
	"sync"
)

type ShortURL string
type LongURL string

type URLStore struct {
	urls map[ShortURL]LongURL
	mu   sync.RWMutex
}

// Gets the LongURL for a given ShortURL, or "" if not found
func (s *URLStore) Get(key ShortURL) LongURL {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.urls[key]
}

// Adds a new mapping, succeeding only if the ShortURL doesn't already exist
func (s *URLStore) Set(key ShortURL, url LongURL) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, present := s.urls[key]
	if present {
		return false
	}
	s.urls[key] = url
	return true
}

// Factory method for creating a URLStore
func NewURLStore() *URLStore {
	return &URLStore{urls: make(map[ShortURL]LongURL)}
}

func (s *URLStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}

func (s *URLStore) Put(url LongURL) ShortURL {
	for {
		key := genKey(s.Count())
		if s.Set(key, url) {
			return key
		}
	}
	// shouldn't get here
	return ""
}

func genKey(n int) ShortURL {
	return ShortURL(fmt.Sprint(md5.Sum([]byte{byte(n)}))[:6])
}
