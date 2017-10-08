package main

import (
	"log"
	"net/rpc"
)

type ProxyStore struct {
	urls   *URLStore
	client *rpc.Client
}

func NewProxyStore(addr string) *ProxyStore {
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		log.Println("Error constructing ProxyStore", err)
	}
	return &ProxyStore{urls: NewURLStore(""), client: client}
}

func (s *ProxyStore) Get(key *ShortURL, url *LongURL) error {
	if err := s.urls.Get(key, url); err == nil { // url foud in local map
		return nil
	}
	// url not found, make RPC call
	if err := s.client.Call("Store.Get", key, url); err != nil {
		return err
	}
	s.urls.Set(key, url) // cache result
	return nil
}

func (s *ProxyStore) Put(url *LongURL, key *ShortURL) error {
	if err := s.client.Call("Store.Put", url, key); err != nil {
		// error doing RPC Put
		return err
	}
	s.urls.Set(key, url)
	return nil
}
