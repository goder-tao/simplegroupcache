package group

import (
	"fmt"
	"testing"
)

func TestGet(t *testing.T)  {
	var db = map[string]string {
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}

	m := NewMember("score", 2<<10, GetFunc(
		func(key string) ([]byte, error) {
			fmt.Println("db hit")
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("no such key data: %v", key)
		}), nil)

	for k,v := range db {
		if bv, err := m.Get(k); err != nil || bv.String() != v {
			t.Error("fail to get k-v")
		}
	}

	if bv, err := m.Get("un"); err == nil || bv.Len() != 0 {
		t.Error("unknown key mustn't get")
	}
}
