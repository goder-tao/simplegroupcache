package test

import (
	"fmt"
	"log"
	"net/http"
	"simplecache/util"
	"testing"
	"time"
)
import "simplecache/groupcache/register"

func TestRegisterFunc(t *testing.T) {
	url := "http://localhost:12301"
	ipport := url[7:]
	// addr := strings.Split(ipport, ":")[0]

	reg := register.NewRegisterCenter(url, time.Second*2)
	go func() {
		log.Fatal(http.ListenAndServe(ipport, reg).Error())
	}()
	fmt.Println("start reg.....")
	// 1.测试注册是否成功
	register.RegisterAndHeartBeat(url, "192.168.50.100", "", time.Second*10)
	resp, err := http.Get(url)
	if err != nil {
		t.Error("get from register: "+err.Error())
	} else {
		cacheServers := resp.Header.Get(register.SERVER)
		util.Info("get server from register: "+cacheServers)
	}
	// 2.过期是否正常
	time.Sleep(time.Second*5)
	resp, err = http.Get(url)
	if err != nil {
		t.Error("expired get from register: "+err.Error())
	} else {
		cacheServers := resp.Header.Get(register.SERVER)
		util.Info("expired get server from register: "+cacheServers)
	}
}
