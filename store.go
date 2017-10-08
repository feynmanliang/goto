package main

import (
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"io"
	"log"
	"os"
	"sync"
)

const saveQueueLength = 1000

type ShortURL string
type LongURL string

type record struct {
	Key ShortURL
	URL LongURL
}

type URLStore struct {
	urls map[ShortURL]LongURL
	mu   sync.RWMutex
	save chan record
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
func NewURLStore(filename string) *URLStore {
	s := &URLStore{
		save: make(chan record, saveQueueLength),
		urls: make(map[ShortURL]LongURL),
	}
	if err := s.load(); err != nil {
		log.Println("Error loading data in URLStore:", err)
	}
	go s.saveLoop(filename)
	return s
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
			s.save <- record{key, url}
			return key
		}
	}
	panic("shouldn't get here")
}

func genKey(n int) ShortURL {
	hasher := md5.New()
	hasher.Write([]byte{byte(n)})
	return ShortURL(hex.EncodeToString(hasher.Sum(nil))[:6])
}

func (s *URLStore) load() error {
	if _, err := s.file.Seek(0, 0); err != nil {
		return err
	}
	d := gob.NewDecoder(s.file)
	var err error
	for err == nil {
		var r record
		if err = d.Decode(&r); err == nil {
			s.Set(r.Key, r.URL)
		}
	}
	if err == io.EOF {
		return nil
	}
	return err
}

func (s *URLStore) saveLoop(filename string) {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("URLStore:", err)
	}
	defer f.Close()
	e := gob.NewEncoder(f)
	for {
		r := <-s.save
		if err := e.Encode(r); err != nil {
			log.Println("URLStore:", err)
		}
	}
}
