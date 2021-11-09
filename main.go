package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"simplecache/groupcache/group"
	HTTP "simplecache/groupcache/http"
	"simplecache/groupcache/register"
	"strconv"
	"time"
)

var db = map[string]string {
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// startCacheService
func startCacheService(resiger, addr string, m *group.Member)  {
	pool := HTTP.NewHTTPPool(addr)
	m.RegisterPeers(pool)
	addrs, err := register.RegisterAndHeartBeat(resiger, addr, 0)
	if err != nil {
		fmt.Println("register and start heartbeat: "+err.Error())
		return
	}
	fmt.Printf("get available server: %v",addrs)
	pool.Set(addrs...)
	log.Println("cache service is starting at "+addr)
	log.Fatal(http.ListenAndServe(addr[7:], pool).Error())
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

	log.Println("font-end service is running at "+apiAddr)
	// handle定义了，handler不必再传值
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

// startRegisterService 开启一个注册中心服务
func startRegisterService(addr string, interval time.Duration)  {
	c := register.NewRegisterCenter(addr, interval)
	log.Println("register service is running at: "+addr)
	log.Fatal(http.ListenAndServe(addr[7:], c))
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

	addrApi := "http://localhost:12000"
	registerAddr := "http://localhost:12001"


	if rgster {
		go startRegisterService(registerAddr, 0)
	}

	if api {
		go startApiService(addrApi, m)
		time.Sleep(1000)
	}

	if len(ip) != 0 {
		startCacheService(registerAddr, "http://"+ip+":"+strconv.Itoa(port), m)
	} else {
		log.Fatal("invalid addr")
	}
}