package lru

import "testing"

type String string

func (s String) Len() int {
	return len(s)
}

func TestAdd(t *testing.T) {
	lru := New(int64(500), nil)

	keys := []string{"key1", "key2", "key3"}
	values := []string{"v1", "v2", "v3"}

	for i := 0; i < len(keys); i++ {
		lru.Add(keys[i], String(values[i]))
	}

	for i := 0; i < len(keys); i++ {
		if v, err := lru.Get(keys[i]); err != nil || v.(String) != String(values[i]) {
			t.Error("not equal")
		}
	}

}

func TestRemove(t *testing.T) {
	lru := New(int64(500), nil)

	keys := []string{"key1", "key2", "key3"}
	values := []string{"v1", "v2", "v3"}

	for i := 0; i < len(keys); i++ {
		lru.Add(keys[i], String(values[i]))
	}

	l1, l2 := lru.ll.Len(), len(lru.cache)

	lru.Remove()

	l11, l22 := lru.ll.Len(), len(lru.cache)

	if l11 != l1-1 || l22 != l2-1 {
		t.Error("remove fail, length error")
	}

	if v, _ := lru.Get(keys[0]); v != nil {
		t.Error("remove k-v fail")
	}
}
