package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"simplecache/group"
	HTTP "simplecache/http"
	"strconv"
)

var db = map[string]string {
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// startCacheService
func startCacheService(addr string, addrs []string, m *group.Member)  {
	pool := HTTP.NewHTTPPool(addr)
	pool.Set(addrs...)
	m.RegisterPeers(pool)
	log.Println("cache service is starting at "+addr)
	log.Fatal(http.ListenAndServe(addr, pool))
}

// startApiService
func startApiService(apiAddr string, m *group.Member) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := m.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "octet-stream")
				w.Write(view.ByteSlice())
			}
		}))

	log.Println("font-end service is running at "+apiAddr)
	// handle定义了，handler不必再传值
	log.Fatal(http.ListenAndServe(apiAddr, nil))
}

func main() {
	var port int
	var api bool

	flag.IntVar(&port, "port", 8001, "cache service port")
	flag.BoolVar(&api, "api", false, "whether to start api server")
	flag.Parse()

	m := group.NewMember("score", 2<<10, group.GetFunc(
		func(key string) ([]byte, error) {
			fmt.Println("db hit")
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("no such key data: %v", key)
		}), nil)

	addrApi := "localhost:9999"
	var addrs  = []string{"http://localhost:8001", "http://localhost:8002", "http:localhost:8003"}

	if api {
		startApiService(addrApi, m)
	}
	startCacheService("http://localhost:"+strconv.Itoa(port), addrs, m)
}