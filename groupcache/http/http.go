// @Description: 负责节点间的通信
// @Author: tao
// @Update: 2021/10/10 14:00

package HTTP

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"simplecache/groupcache/consistenthash"
	"simplecache/groupcache/group"
	"simplecache/groupcache/lru"
	"simplecache/groupcache/pb"
	"simplecache/groupcache/peer"
	"strconv"
	"strings"
	"sync"
)


const (
	// 域名下用来做缓存之间通信的子路径
	defaultBasePath = "/simplecache/"
	defaultReplicas = 5
	LISTEN_PORT = 13000
)

// 1. 负责节点间通信的数据结构, 每个节点持有一个，保存所有获取到其他节点的信息和方式
// 2. 完成节点的选择（一致性hash）
type HTTPPool struct {
	ipport        string
	basePath      string
	mu            sync.Mutex
	peerMap       *consistenthash.Map
	httpGetterMap map[string]*httpGetter // key ---> http://ip:port
}

// httpGetter 负责使用http从peer获得key对应数据
type httpGetter struct {
	// 远程节点的域名地址/basePath/
	baseURL string
}
var _ peer.PeerGetter = (*httpGetter)(nil)

func NewHTTPPool(ipport string) *HTTPPool {
	return &HTTPPool{
		ipport:   ipport,
		basePath: defaultBasePath,
	}
}

// <--- HTTPPool方法 --->
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.ipport, fmt.Sprintf(format, v...))
}

func (p HTTPPool)GetAddr() string {
	return p.ipport
}

// Set init or reset the pool member
func (p *HTTPPool) Set(urls ...string)  {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peerMap = consistenthash.New(defaultReplicas, nil)
	p.peerMap.Add(urls...)
	p.httpGetterMap = make(map[string]*httpGetter)
	for _, u := range urls {
		p.httpGetterMap[u] = &httpGetter{baseURL: "http://"+u+strconv.Itoa(LISTEN_PORT)+p.basePath}
	}
}

// Pick pick the specific httpGetter
func (p *HTTPPool) Pick(key string) (peer.PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if addr := p.peerMap.Get(key); addr != "" && addr != p.ipport {
		p.Log("pick peer %v", addr)
		return p.httpGetterMap[addr], true
	} else {
		log.Println("pick self: "+p.ipport)
	}
	return nil, false
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	servers := r.Header.Get("Servers")
	if servers != "" {
		switch r.Method {
		case "POST":
			log.Println("new peer:"+servers)
			p.peerMap.Add(strings.Split(servers, ",")...)
			for _, u := range strings.Split(servers, ",") {
				p.httpGetterMap[u] = &httpGetter{baseURL: "http://"+u+strconv.Itoa(LISTEN_PORT)+p.basePath}
			}
		case "DELETE":
			// bug
			p.peerMap.Delete(strings.Split(servers, ",")...)
			for _, u := range strings.Split(servers, ",") {
				delete(p.httpGetterMap, u)
			}
		default:
		}
	} else {
		p.servePeer(w, r)
	}
}

func (p *HTTPPool) servePeer(w http.ResponseWriter, r *http.Request)  {
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

	view, err := m.PeerGet(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, "proto encoding fail", http.StatusInternalServerError)
	}

	// 几种常见的http content-type
	// 1. 媒体格式：
	//    text/html
	//    text/xml
	//    text/gif
	//    text/png
	//    text/jpeg
	// 2. application格式:
	//	  application/xml
	//	  application/pdf
	//	  application/json
	//	  application/octet-stream  --- 二进制流数据
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

// <--- httpGetter的方法 --->
func (h *httpGetter) Get(name, key string) (lru.ByteValue, error) {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(name), url.QueryEscape(key))

	res, err := http.Get(u)
	if err != nil {
		return lru.ByteValue{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return lru.ByteValue{}, errors.New("get fail, status code: "+string(res.StatusCode))
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return lru.ByteValue{}, errors.New("reading response body: "+err.Error())
	}

	p := &pb.Response{}
	if err := proto.Unmarshal(bytes, p); err != nil {
		return lru.ByteValue{}, err
	}

	return lru.ByteValue{B: p.Value}, nil
}