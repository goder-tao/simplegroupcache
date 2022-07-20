package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"simplecache/groupcache/group"
	HTTP "simplecache/groupcache/http"
	"simplecache/groupcache/register"
	"simplecache/rpc"
	"simplecache/util"
	"strconv"
	"time"
)

var db = map[string]string {
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// StartCacheServer 可选tcp或者http开启缓存服务
func StartCacheServer(network, ip, registerAddr, codecType string, port int, m *group.Member) {
	switch network {
	case "tcp":
		pool := rpc.NewRPCPool(ip+":"+strconv.Itoa(port+rpc.OFF), nil)
		m.RegisterPeers(pool)
		go pool.AcceptAndServe()
		addrs, codecs, err := register.RegisterAndHeartBeat(registerAddr, ip+":"+strconv.Itoa(port), codecType, 0)
		if err != nil {
			util.Error("register and start heartbeat: "+err.Error())
			return
		}
		if len(codecs) != 0{
			if len(codecs) != len(addrs) {
				util.Fatal("codecs and addrs length no equal")
			} else {
				for i := 0 ; i < len(codecs); i++ {
					pool.Add(codecs[i], addrs[i])
				}
			}
		} else {
			pool.Add(rpc.DEFAULT_CODEC, addrs...)
		}
		util.Info("cache service is starting at "+ip+":"+strconv.Itoa(port+123))
		// http监听注册中心的消息
		util.Fatal(http.ListenAndServe(ip+":"+strconv.Itoa(port), pool).Error())
	case "http":
		pool := HTTP.NewHTTPPool(ip+":"+strconv.Itoa(port))
		m.RegisterPeers(pool)
		addrs, _, err := register.RegisterAndHeartBeat(registerAddr, ip+":"+strconv.Itoa(port), "", 0)
		if err != nil {
			util.Info("register and start heartbeat: "+err.Error())
			return
		}
		pool.Set(addrs...)
		util.Info("cache service is starting at "+ip+":"+strconv.Itoa(port))
		util.Fatal(http.ListenAndServe(ip+":"+strconv.Itoa(port), pool).Error())
	default:
		util.Fatal("unknown network")
		return
	}
}


// startApiService 外露一个api接口，让所有的前端请求都打到这台机器上，api服务器上先查看是否
// 请求的数据缓存在本地，如果不是调用concurrenthahs找到目标
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

	util.Info("font-end service is running at "+apiAddr)
	// handle定义了，handler不必再传值
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

// startRegisterService 开启一个注册中心服务
func startRegisterService(url string, interval time.Duration)  {
	c := register.NewRegisterCenter(url, interval)
	util.Info("register service is running at: "+url)
	util.Error(http.ListenAndServe(url[7:], c).Error())
}

func main() {
	var ip string
	var port int
	var api bool
	var rgster bool

	flag.IntVar(&port, "port", 8001, "cache service port")
	flag.BoolVar(&api, "api", false, "whether to start api server")
	flag.StringVar(&ip, "ip", "", "cache server ip")
	flag.BoolVar(&rgster, "register", false, "whether to start register")
	flag.Parse()

	m := group.NewMember("score", 2<<10, group.GetFunc(
		func(key string) ([]byte, error) {
			fmt.Println("db hit")
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("no such key data: %v", key)
		}), nil)

	// addrApi := "http://localhost:12000"
	registerAddr := "http://localhost:12012"


	if rgster {
		go startRegisterService(registerAddr, 0)
	}

	//if api {
	//	go startApiService(addrApi, m)
	//	time.Sleep(time.Second)
	//}

	if len(ip) != 0 {
		time.Sleep(1*time.Second)
		StartCacheServer("tcp", ip, registerAddr, rpc.DEFAULT_CODEC, port, m)
	} else {
		log.Fatal("invalid addr")
	}

}