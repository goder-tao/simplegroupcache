// @Description: 负责节点间的通信
// @Author: tao
// @Update: 2021/10/10 14:00
package HTTP

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"simplecache/group"
)

const defaultBasePath = "/simplecache/"

// 节点间通信的数据结构
type HTTPPool struct {
	self string
	basePath string
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self: self,
		basePath: defaultBasePath,
	}
}

// <--- HTTPPool方法 --->
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

//
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// url必须以basePath开头
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		http.Error(w, "HTTPPool server unexpected path: "+r.URL.Path, http.StatusForbidden)
		return
	}

	p.Log("%s %s", r.Method, r.URL.Path)

	// url格式: /basePath/memberName/key
	paths := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)

	if len(paths) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	memberName := paths[0]
	key := paths[1]

	m := group.GetMember(memberName)
	if m == nil {
		http.Error(w, "no such member: "+memberName, http.StatusNotFound)
		return
	}

	view, err := m.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 几种常见的http content-type
	// 1. 媒体格式：
	//    text/html
	//    text/xml
	//    text/gif
	//    text/png
	//    text/jpeg
	// 2. application格式
	//	  application/xml
	//	  application/pdf
	//	  application/json
	//	  application/octet-stream  --- 二进制流数据
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}