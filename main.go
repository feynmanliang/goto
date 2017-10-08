package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/rpc"
)

var (
	listenAddr = flag.String("http", ":8080", "http listen address")
	dataFile   = flag.String("file", "store.json", "data store file name")
	hostname   = flag.String("host", "localhost:8080", "host name and port")
	rpcEnabled = flag.Bool("rpc", false, "enable RPC server")
	masterAddr = flag.String("master", "", "RPC master address")
)

var store Store

func main() {
	flag.Parse()
	if *masterAddr != "" { // we are a slave
		store = NewProxyStore(*masterAddr)
	} else { // we are the master
		store = NewURLStore(*dataFile)
	}
	if *rpcEnabled {
		rpc.RegisterName("Store", store)
		rpc.HandleHTTP()
	}
	http.HandleFunc("/", Redirect)
	http.HandleFunc("/add", Add)
	http.ListenAndServe(*listenAddr, nil)
}

func Add(w http.ResponseWriter, r *http.Request) {
	url := LongURL(r.FormValue("url"))
	if url == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, string(AddForm))
		return
	}
	var key ShortURL
	if err := store.Put(&url, &key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "http://%s/%s", *hostname, key)
}

const AddForm = `
<form method="POST" action="/add">
URL: <input type="text" name="url">
<input type="submit" value="Add">
</form>
`

func Redirect(w http.ResponseWriter, r *http.Request) {
	key := ShortURL(r.URL.Path[1:])
	var url LongURL
	if err := store.Get(&key, &url); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, string(url), http.StatusFound)
}
