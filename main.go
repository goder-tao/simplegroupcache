package main

import (
	"fmt"
	"log"
	"net/http"
	"simplecache/group"
	HTTP "simplecache/http"
)

var db = map[string]string {
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	group.NewMember("score", 2<<10, group.GetFunc(
		func(key string) ([]byte, error) {
			fmt.Println("db hit")
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("no such key data: %v", key)
		}), nil)

	addr := "localhost:9999"
	pool := HTTP.NewHTTPPool(addr)
	log.Println("cache is running at "+addr)

	log.Fatal(http.ListenAndServe(addr, pool))
}
