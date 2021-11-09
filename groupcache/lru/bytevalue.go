// @Description: 保存cache的value的实际对象
// @Author: tao
// @Date: 2021/10/08 17:21
package lru

type ByteValue struct {
	B []byte
}

func (bv ByteValue) Len() int {
	return len(bv.B)
}

func (bv ByteValue) String() string {
	return string(bv.B)
}

// 提供只读，防止修改cache中的数据
func (bv ByteValue) ByteSlice() []byte {
	bb := make([]byte, len(bv.B))
	copy(bb, bv.B)
	return bb
}
