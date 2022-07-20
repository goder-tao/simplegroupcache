package bloomfilter

import (
	"crypto/rand"
	"encoding/binary"
	"hash"
	"log"
	"simplecache/groupcache/filter"
)


// bloomFilter 利用一个hash+多个随机key的位操作组合产生不同的hash值，
// hash 根据key不同产生结果不同
type bloomFilter struct {
	N, K int
	bitmap []uint64
	rKeys []uint64
	hash hash.Hash64
}

var _ filter.Filter = (*bloomFilter)(nil)

func (b *bloomFilter) Put(key string) {
	b.hash.Reset()
	_, _ =b.hash.Write([]byte(key))
	rawHash := b.hash.Sum64()
	for _, k := range b.rKeys {
		b.setBit((rawHash^k)%uint64(b.N))
	}
}

func (b *bloomFilter) Contain(key string) bool {
	b.hash.Reset()
	_, _ = b.hash.Write([]byte(key))
	rawHash := b.hash.Sum64()
	for _, k := range b.rKeys {
		if !b.isBitSet((rawHash^k)%uint64(b.N)) {
			return false
		}
	}
	return true
}

// setBit set specific pos of bitmap
func (b *bloomFilter) setBit(pos uint64)  {
	b.bitmap[pos>>6] |= 1 << (pos%64)
}

func (b *bloomFilter) isBitSet(pos uint64) bool {
	return (b.bitmap[pos>>6] & (1<<(pos%64))) != 0
}

func randKeys(K int) []uint64 {
	keys := make([]uint64, K)
	err := binary.Read(rand.Reader, binary.LittleEndian, keys)
	if err != nil {
		log.Panicf(
			"Cannot read %d bytes from CSRPNG crypto/rand.Read (err=%v)",
			8, err,
		)
	}
	return keys
}

func New(N, K int, hash hash.Hash64) filter.Filter {
	return &bloomFilter{
		N: N,
		K: K,
		rKeys: randKeys(K),
		bitmap: make([]uint64, (N+63)/64),
		hash: hash,
	}
}