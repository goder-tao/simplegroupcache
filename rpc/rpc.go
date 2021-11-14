package rpc

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"simplecache/groupcache/consistenthash"
	"simplecache/groupcache/group"
	"simplecache/groupcache/lru"
	"simplecache/groupcache/peer"
	"simplecache/groupcache/register"
	"simplecache/rpc/codec"
	"simplecache/util"
	"strconv"
	"strings"
	"sync"
)

const (
	replicas = 5
	RPC_PREFIX = "/__rpc__/"
	DEFAULT_CODEC = codec.GOB
	OFF = 123
)

// RPCPool 实现了PeerPicker接口，提供http调用的功能
type RPCPool struct {
	rw sync.RWMutex
	addr string
	csHash *consistenthash.Map
	peerMap map[string]*rpcGetter
	off int
}

// 注册中心通信
func (R *RPCPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	servers := r.Header.Get(register.SERVER)
	if servers != "" {
		codecType := r.Header.Get(register.CODEC)
		if len(codecType) == 0 {
			codecType = DEFAULT_CODEC
		}
		switch r.Method {
		case "POST":
			util.Info("new peer:"+servers)
			for _, u := range strings.Split(servers, ",") {
				pu := strings.Split(u, ":")
				u = pu[0]+":"
				port, _ := strconv.Atoi(pu[1])
				ps := strconv.Itoa(port+OFF)
				u = u+ps
				R.csHash.Add(u)
				R.peerMap[u] = &rpcGetter{dialURL: u, codecType: codecType,}
			}
		case "DELETE":
			// bug
			R.csHash.Delete(strings.Split(servers, ",")...)
			for _, u := range strings.Split(servers, ",") {
				delete(R.peerMap, u)
			}
		default:
		}
	}
}

func NewRPCPool(addr string, hashFunc consistenthash.Hash) *RPCPool {
	return &RPCPool{
		addr: addr,
		csHash: consistenthash.New(replicas, hashFunc),
		peerMap: make(map[string]*rpcGetter),
		off: OFF,
	}
}

func (R *RPCPool) GetAddr() string {
	return R.addr
}

// acceptAndServe 开启一个listener持续接收conn，再开启子线程处理conn
func (R *RPCPool) AcceptAndServe() {
	nl, err := net.Listen("tcp", R.addr)
	if err != nil {
		log.Fatal("listen net "+err.Error())
		return
	}
	for {
		conn, err := nl.Accept()
		if err != nil{
			log.Println("listener accept "+err.Error())
			continue
		}
		go R.serveConn(conn)
	}
}

func (R *RPCPool) serveConn(conn net.Conn) {
	infoH := codec.InfoHeader{}
	if err := json.NewDecoder(conn).Decode(&infoH); err != nil {
		log.Println("info header decode "+err.Error())
		return
	}
	cdc := codec.CodecFunMap[infoH.GobType](conn)
	for {
		var header codec.RequestHeader
		// 解码header
		log.Println("000")
		err := cdc.ReadHeader(&header)
		log.Println("111")
		if err != nil {
			header.Err = "internal error "+err.Error()
			err = cdc.Write(&header, 0)
			if err != nil {
				util.Error("internal serve con "+err.Error())
			}
			err = cdc.Close()
			if err != nil {
				util.Error("internal close conn "+err.Error())
			}
			break
		}
		// 获取cache并尝试从当前cache server获取数据
		cache := group.GetMember(header.Name)
		value, err := cache.PeerGet(header.Key)
		log.Println("222")
		if err != nil {
			header.Err = "internal error "+err.Error()
			err = cdc.Write(&header, 0)
			if err != nil {
				util.Error("internal serve con "+err.Error())
			}
			err = cdc.Close()
			if err != nil {
				util.Error("internal close conn "+err.Error())
			}
			break
		}
		// 数据写回
		err = cdc.Write(&header, value.ByteSlice())
		log.Println(2)
		if err != nil {
			util.Error("write response "+err.Error())
			err = cdc.Close()
			if err != nil {
				util.Error("internal close conn "+err.Error())
			}
			break
		}
	}
}

// Pick 接口函数，使用哈希一致性获取peer
func (R *RPCPool) Pick(key string) (peer peer.PeerGetter, ok bool) {
	R.rw.RLock()
	defer R.rw.RUnlock()
	peerAddr := R.csHash.Get(key)
	if peer, ok = R.peerMap[peerAddr]; !ok {
		log.Fatalf("no peer %v\n", peerAddr)
		return nil, false
	} else {
		if peerAddr == R.addr {
			util.Info("pick self "+peerAddr)
		} else {
			util.Info("pick peer "+peerAddr)
		}
		return
	}
}

// Add 添加一系列的peer到RPCPool中
func (R *RPCPool) Add(codecType string, peerAddr ...string) {
	R.rw.Lock()
	defer R.rw.Unlock()
	R.csHash.Add(peerAddr...)
	for _, addr := range peerAddr {
		R.peerMap[addr] = &rpcGetter{
			dialURL: addr,
			codecType: codecType,
		}
	}
}

// rpcGetter 一个向其他所有peer发送rpc请求并接收恢复的rpc客户端
type rpcGetter struct {
	mu sync.Mutex
	// tcp dial的路径
	dialURL string
	codecType string
	// 同一个getter使用的codec应该相同
	cdc codec.Codec
	// call的序列，标识不同的call
	isOpened bool
	sqe int
	// 可能向多个cache server同时发送请求
}

var _ peer.PeerGetter = (*rpcGetter)(nil)

// Get 实现利用rpc调用远程peer的功能
func (rg *rpcGetter) Get(name, key string) (lru.ByteValue, error) {
	// 保证一次只有一个流发送
	rg.mu.Lock()
	defer rg.mu.Unlock()

	// 先初始化codec
	if rg.cdc == nil || !rg.isOpened {
		conn, err :=net.Dial("tcp", rg.dialURL)
		if err != nil {
			return lru.ByteValue{}, err
		}
		// 先发送一个InfoHeader
		infoH := codec.InfoHeader{GobType: rg.codecType}
		if err = json.NewEncoder(conn).Encode(&infoH); err != nil {
			return lru.ByteValue{}, err
		}
		rg.cdc = codec.CodecFunMap[rg.codecType](conn)
		rg.isOpened = true
	}

	log.Println("0000")
	req := codec.RequestHeader{Name: name, Key: key}
	// 发送header等待接收回复
	if err := rg.cdc.Write(&req, 0); err != nil {
		return lru.ByteValue{}, err
	}
	log.Println("0111")
	var data []byte
	header := codec.RequestHeader{}
	if err := rg.cdc.ReadHeader(&header); err != nil {
		return lru.ByteValue{}, err
	}
	log.Println("0222")
	if header.Err != "" {
		rg.isOpened = false
		return lru.ByteValue{}, errors.New(header.Err)
	}
	log.Println("0333")
	err := rg.cdc.ReaderBody(&data)
	return lru.ByteValue{B: data}, err
}

// receive 开启线程程尝试一直接收从其他peer的tcp回复
func (rg *rpcGetter)receive() {
	var err error
	for err == nil {

	}
}
