package main

type Store interface {
	Put(url *LongURL, key *ShortURL) error
	Get(key *ShortURL, url *LongURL) error
}
