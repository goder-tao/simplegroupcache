// @author: tao
// @data: 2021-11-8

// center offer simple server register and heartbeat mechanism
package register

import (
	"log"
	"net/http"
	"simplecache/util"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DEFAULT_CHECK_INTERVAL = time.Second * 20
	DEFAULT_HEARTBEAT_INTERVAL = time.Second * 10

	// code
	REGISTER = "Register"
	SERVER = "Server"
	CODEC = "Codec"
)

// Center 是注册中心的主要数据结构
type Center struct {
	addr string
	mu sync.Mutex
	servers map[string]*serverItem
	checkInterval time.Duration  // 注册中心健康检查时间
}

// serverItem 记录每个注册的server的地址和上次更新的时间
type serverItem struct {
	Addr string
	CodecType string
	lastUpdateTIme time.Time
}

func NewRegisterCenter(addr string, checkInterval time.Duration) *Center {
	if checkInterval == 0 {
		checkInterval = DEFAULT_CHECK_INTERVAL
	}
	c := &Center{
		servers:  make(map[string]*serverItem),
		checkInterval: checkInterval,
		addr: addr,
	}

	go func() {
		t := time.NewTicker(c.checkInterval)
		for {
			<- t.C
			c.healthyCheck()
		}
	}()
	return c
}

// 注册中心以http的方式和每个server通信
// POST: 注册注册或者是心跳
// GET: 获取所有存活的服务器
func (c *Center) ServeHTTP(rsp http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		addr := req.Header.Get(REGISTER)
		codecType := req.Header.Get(CODEC)
		if addr == "" {
			util.Error("register receive empty address")
			rsp.WriteHeader(http.StatusBadRequest)
		} else {
			c.register(addr, codecType)
		}
	case "GET":
		rsp.Header().Set(SERVER, strings.Join(c.getAvailableServer(), ","))
		rsp.Header().Set(CODEC, strings.Join(c.getAllCodec(), ","))
	default:
		rsp.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// register 注册一个server或者是更新server的lastUpdate
func (c *Center) register(addr, codecType string) {
	if item, ok := c.servers[addr]; ok {
		item.lastUpdateTIme = time.Now()
	} else {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.servers[addr] = &serverItem{
			Addr: addr,
			CodecType: codecType,
			lastUpdateTIme: time.Now(),
		}
		for addrr, _ := range c.servers {
			if addrr == addr {
				continue
			}
			httpClient := &http.Client{}
			req, _ := http.NewRequest("POST", "http://"+addrr, nil)
			req.Header.Set(SERVER, addr)
			req.Header.Set(CODEC, codecType)
			_, _ = httpClient.Do(req)
		}
		util.Info("register server: "+addr)
	}
	return
}

// healthyCheck 健康检查，在创建注册中心对象的时候开启一个线程持续循环调用
func (c *Center) healthyCheck() {
	unhealth := []string{}
	for addr, item := range c.servers {
		// 超时未心跳
		if time.Now().Nanosecond()-item.lastUpdateTIme.Nanosecond() > int(DEFAULT_CHECK_INTERVAL.Nanoseconds()) {
			unhealth = append(unhealth, item.Addr)
			c.mu.Lock()
			delete(c.servers, addr)
			c.mu.Unlock()
		}
	}

	unhealthJoin := strings.Join(unhealth, ",")
	// 对于不健康的节点通知所有的其他server
	if len(unhealth) > 0 {
		for addr, _ := range c.servers {
			httpClient := &http.Client{}
			req, _ := http.NewRequest("DELETE", addr, nil)
			req.Header.Set(SERVER, unhealthJoin)
			httpClient.Do(req)
		}
	}
}

//
func (c *Center) getAvailableServer() []string {
	servers := []string{}
	for server, _ := range c.servers {
		servers = append(servers, server)
	}
	return servers
}

func (c *Center) getAllCodec() []string {
	cdcs := []string{}
	for _, item := range c.servers {
		cdcs = append(cdcs, item.CodecType)
	}
	return cdcs
}

// RegisterAndHeartBeat register暴露的方法，供cache server创建的时候调用，首先发送一个HeartBeat
// 进行注册，之后再开启一个线程持续间隔发送HeartBeat
func RegisterAndHeartBeat(registerAddr, serverAddr, codecType string, interval time.Duration) ([]string, []string, error) {
	if interval == 0 {
		interval = DEFAULT_HEARTBEAT_INTERVAL
	}

	if err := sendHeartBeat(registerAddr, serverAddr, codecType); err != nil {
		return nil, nil, err
	}

	// 注册成功后获取可用服务列表
	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", registerAddr, nil)
	rsp, err := httpClient.Do(req)

	if err != nil {
		return nil, nil, err
	} else {
		// 持续发送心跳
		go func() {
			t := time.NewTicker(interval)
			for {
				<-t.C
				if err := sendHeartBeat(registerAddr, serverAddr, codecType); err != nil {
					log.Println("send heartbeat fail:"+err.Error())
					break
				}
			}
		}()

		servers := strings.Split(rsp.Header.Get(SERVER), ",")
		for i, addr := range servers {
			paddr := strings.Split(addr, ":")
			port, _ := strconv.Atoi(paddr[1])
			ps := strconv.Itoa(port+123)
			addr = paddr[0]+":"+ps
			servers[i]=addr
		}
		codecs := strings.Split(rsp.Header.Get(CODEC), ",")
		return servers, codecs, nil
	}
}

// 发送心跳
func sendHeartBeat(register, server, codecType string) error {
	httpClient := &http.Client{}
	req, err := http.NewRequest("POST", register, nil)
	if err != nil {
		return err
	}
	req.Header.Set(REGISTER, server)
	req.Header.Set(CODEC, codecType)
	_, err = httpClient.Do(req)
	return err
}