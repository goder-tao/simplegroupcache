package test

import (
	"errors"
	"fmt"
	"net/http"
	"simplecache/groupcache/group"
	Http "simplecache/groupcache/http"
	"simplecache/groupcache/lru"
	"simplecache/groupcache/register"
	"simplecache/util"
	"testing"
	"time"
)

var db = map[string]string{
	"Tony": "13",
	"Bob": "14",
	"Ming": "20",
	"Sam": "25",
}

func TestCacheServer(t *testing.T)  {
	regIp := "localhost"
	regPort := 12000
	cachePort := 13000
	dataType := "person"
	cacheSize := int64(1024)
	codecType := ""
	beatInterval := 3

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
		t.Error("RegisterAndHeartBeat failed, "+err.Error())
	}
	picker.Set(addrs...)

	util.Fatal(http.ListenAndServe(ipport, picker).Error())
	// test point


}
