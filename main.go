package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type cache struct {
	mu sync.Mutex
	m  map[string][]byte
}

func newCache() *cache {
	return &cache{
		m: make(map[string][]byte),
	}

}

func (c *cache) Add(k string, b []byte) {
	c.mu.Lock()
	c.m[k] = b
	c.mu.Unlock()
}

func (c *cache) Get(k string) []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m[k]
}

func (c *cache) Keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := make([]string, 0, len(c.m))
	for k := range c.m {
		keys = append(keys, k)
	}
	return keys
}

type server struct {
	addr  string
	cache *cache
	r     *mux.Router
}

func newServer(addr string, c *cache) *server {
	r := mux.NewRouter()

	srv := server{
		addr:  addr,
		r:     r,
		cache: c,
	}
	srv.r.HandleFunc("/junk/{key}", srv.get)
	srv.r.HandleFunc("/", srv.post).Methods("POST")
	srv.r.HandleFunc("/", srv.index)
	return &srv
}

func (s *server) get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	if key == "" {
		http.Error(w, "missing key", 400)
		return
	}

	body := s.cache.Get(key)
	if len(body) == 0 {
		http.Error(w, "unknown junk", 400)
		return
	}

	w.Write(body)
}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "welcome to the junkyard\n")
}

func (s *server) post(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("error reading body:%s", err), 500)
		return
	}

	hash := fmt.Sprintf("%x", md5.Sum(body))
	s.cache.Add(hash, body)
	path := fmt.Sprintf("%s/junk/%s\n", r.Host, hash)
	io.WriteString(w, path)
}

func (s *server) ListenAndServe() error {
	log.Printf("running on %s", s.addr)
	logged := handlers.LoggingHandler(os.Stderr, s.r)
	return http.ListenAndServe(s.addr, logged)
}

func main() {
	c := newCache()
	addr := ":3000"
	srv := newServer(addr, c)
	log.Fatal(srv.ListenAndServe())
}
