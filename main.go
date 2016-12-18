package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type server struct {
	addr  string
	cache *cache
	mux   *http.ServeMux
}

func newServer(addr string, c *cache) *server {
	mux := http.NewServeMux()

	srv := server{
		addr:  addr,
		mux:   mux,
		cache: c,
	}
	mux.HandleFunc("/", srv.index)
	return &srv
}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		s.get(w, r)
	case r.Method == "POST":
		s.post(w, r)
	default:
		http.Error(w, "bad request", 400)
	}
}

func (s *server) get(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		io.WriteString(w, os.Args[0]+"\n")
		return
	}

	key := r.URL.Path[1:]
	body := s.cache.Get(key)
	if len(body) == 0 {
		http.Error(w, "not found", 404)
		return
	}

	// FIXME catch errors
	w.Write(body)
}

func (s *server) post(w http.ResponseWriter, r *http.Request) {
	// FIXME permit max size
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("error reading body:%s", err), 500)
		return
	}

	hash := fmt.Sprintf("%x", md5.Sum(body))
	s.cache.Add(hash, body)
	path := fmt.Sprintf("%s/%s\n", r.Host, hash)
	io.WriteString(w, path)
}

func (s *server) ListenAndServe() error {
	log.Printf("running on %s", s.addr)
	return http.ListenAndServe(s.addr, s.mux)
}

func main() {
	size := 1000
	expiry := time.Minute * 15
	c := newCache(size, expiry)
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	addr := ":" + port
	srv := newServer(addr, c)
	log.Fatal(srv.ListenAndServe())
}
