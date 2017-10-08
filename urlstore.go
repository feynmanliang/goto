package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
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

// Factory method for creating a URLStore
func NewURLStore(filename string) *URLStore {
	s := &URLStore{
		urls: make(map[ShortURL]LongURL),
	}
	if filename != "" {
		s.save = make(chan record, saveQueueLength)
		if err := s.load(filename); err != nil {
			log.Println("Error loading URLStore: ", err)
		}
		go s.saveLoop(filename)
	}
	return s
}

// Gets the LongURL for a given ShortURL, or "" if not found
func (s *URLStore) Get(key *ShortURL, url *LongURL) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if u, ok := s.urls[*key]; ok {
		*url = u
		return nil
	}
	return errors.New("key not found")
}

// Generates a ShortURL for a LongURL and Sets the entry in the store
func (s *URLStore) Put(url *LongURL, key *ShortURL) error {
	for {
		*key = genKey(s.Count())
		if err := s.Set(key, url); err == nil {
			break
		}
	}
	if s.save != nil {
		s.save <- record{*key, *url}
	}
	return nil
}

// Adds a new mapping, succeeding only if the ShortURL doesn't already exist
func (s *URLStore) Set(key *ShortURL, url *LongURL) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, present := s.urls[*key]; present {
		return errors.New("key already exists")
	}
	s.urls[*key] = *url
	return nil
}

func (s *URLStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}

func genKey(n int) ShortURL {
	hasher := md5.New()
	hasher.Write([]byte{byte(n)})
	return ShortURL(hex.EncodeToString(hasher.Sum(nil))[:6])
}

func (s *URLStore) load(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}
	d := json.NewDecoder(file)
	for err == nil {
		var r record
		if err = d.Decode(&r); err == nil {
			s.Set(&r.Key, &r.URL)
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
	e := json.NewEncoder(f)
	for {
		r := <-s.save
		if err := e.Encode(r); err != nil {
			log.Println("URLStore:", err)
		}
	}
}
