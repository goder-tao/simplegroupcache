package bloomfilter

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"testing"
)



func TestBloomFilter(t *testing.T)  {
	bf := New(100000, 3, fnv.New64())
	testN := 10000

	for i := 0; i < testN; i++ {
		bf.Put(strconv.Itoa(i))
	}

	errN := 0
	for i := 0; i < testN; i++ {
		if !bf.Contain(strconv.Itoa(i)) {
			errN += 1
		}
	}

	fmt.Println(fmt.Sprintf("[contain test] errN: %d, err rate: %f", errN, float64(errN)/float64(testN)))

	errN = 0
	for i := testN; i < testN*2; i++ {
		if bf.Contain(strconv.Itoa(i)) {
			errN += 1
		}
	}
	fmt.Println(fmt.Sprintf("[not contained test] errN: %d, err rate: %f", errN, float64(errN)/float64(testN)))
}