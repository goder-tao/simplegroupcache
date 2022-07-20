package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"simplecache/groupcache/group"
	Http "simplecache/groupcache/http"
	"simplecache/groupcache/lru"
	"simplecache/groupcache/register"
	"simplecache/util"
	"syscall"
	"time"
)

var db = map[string]string{
	"Tony": "13",
	"Bob": "14",
	"Ming": "20",
	"Sam": "25",
}

func startCacheService(dataType string, cacheSize int64, regIp string, regPort, cachePort int, codecType string,
	beatInterval int64)  {
	regUrl := fmt.Sprintf("http://%s:%d", regIp, regPort)
	ipport := fmt.Sprintf("%s:%d", "localhost", cachePort)
	// 1. 数据源
	getter := group.GetFunc(func(key string) ([]byte, error) {
		util.Info("retrieve data from db")
		if v, ok := db[key]; ok {
			return []byte(v), nil
		} else {
			return nil, errors.New("no suck key in db")
		}
	})
	// 2. 创建一个member
	// 淘汰写回法
	evicted := func(key string, value lru.Value) {
		db[key] = string(value.(lru.ByteValue).ByteSlice())
	}

	memb := group.NewMember(dataType, cacheSize, getter, evicted)
	// 3. 创建一个peerPicker
	picker := Http.NewHTTPPool(ipport)
	memb.RegisterPeers(picker)

	// 4. member注册
	time.Sleep(time.Second)
	addrs, _, err := register.RegisterAndHeartBeat(regUrl, ipport, codecType, time.Duration(beatInterval)*time.Millisecond)
	if err != nil {
		util.Fatal("RegisterAndHeartBeat failed, "+err.Error())
	}
	picker.Set(addrs...)

	util.Fatal(http.ListenAndServe(ipport, picker).Error())
}

func main() {
	var regAddr string
	var regPort int = register.LISTEN_PORT
	var dataType string
	var cacheSize int64
	var serverPort int = Http.LISTEN_PORT
	var codecType string
	var interval int
	flag.StringVar(&regAddr, "rip", "", "register ip addr")
	flag.StringVar(&codecType, "codecType", "", "codec type")
	flag.StringVar(&dataType, "dataType", "", "data type")
	flag.Int64Var(&cacheSize, "cacheSize", 1024, "cache size, unit: Byte")
	flag.IntVar(&interval, "interval", 1000, "server heartbeat interval")
	flag.Parse()

	go func() {
		startCacheService(dataType, cacheSize,regAddr, regPort, serverPort, codecType, int64(interval))
	}()


	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh,syscall.SIGTERM)
	select {
	case <- termCh:
		util.Info("receive SIGTERM, stopping...")
		util.Info("end")
	}

}
