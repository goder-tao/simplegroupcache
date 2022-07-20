package main

import (
	"flag"
	"fmt"
	"net/http"
	"simplecache/groupcache/register"
	"simplecache/util"
	"time"
)

func startRegisterService(port int, interval time.Duration)  {
	c := register.NewRegisterCenter("0.0.0.0", interval)
	ap := fmt.Sprintf("%s:%d", "localhost", port)
	util.Info("register center created, start listen")
	util.Fatal(http.ListenAndServe(ap, c).Error())
}

func main() {
	var interval int64
	flag.Int64Var(&interval, "interval", 5000, "register health check interval, unit: ms")
	flag.Parse()

	startRegisterService(register.LISTEN_PORT, time.Millisecond*time.Duration(interval))
}



