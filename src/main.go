package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

var wg3 sync.WaitGroup
var ch1 = make(chan int, 50)

func main() {

	InitData()
	//开启隧道代理
	go RunTunnelProxyServer()
	//启动webAPi
	Run()

}

// 初始化
func InitData() {
	//获取配置文件
	GetConfigData()
	//是否需要抓代理
	if len(ProxyPool) < conf.Config.ProxyNum {
		//抓取代理
		spiderRun()
		//导出代理到文件
		export()
	}

	//定时判断是否需要获取代理iP
	go func() {
		// 每 60 秒钟时执行一次
		ticker := time.NewTicker(60 * time.Second)
		for range ticker.C {
			if len(ProxyPool) < conf.Config.ProxyNum {
				if !run {
					//抓取代理
					spiderRun()
					//导出代理到文件
					export()
				}

			}
		}
	}()

	//定时更换隧道IP
	go func() {
		tunnelTime := time.Duration(conf.Config.TunnelTime)
		ticker := time.NewTicker(tunnelTime * time.Second)
		for range ticker.C {
			if len(ProxyPool) == 0 {
				continue
			}
			Iip = getIIp()
			Sip = getSIp()
		}
	}()

	// 验证代理存活情况
	go func() {
		verifyTime := time.Duration(conf.Config.VerifyTime)
		ticker2 := time.NewTicker(verifyTime * time.Second)
		for range ticker2.C {
			for i, _ := range ProxyPool {
				ProxyPool[i].RequestNum = 1
				ProxyPool[i].SuccessNum = 1
			}
			for io := 0; io < 4; io++ {
				for i := range ProxyPool {
					wg3.Add(1)
					ch1 <- 1
					Verify(&ProxyPool[i], &wg3, ch1)
				}
			}
			wg3.Wait()
			lock.Lock()
			for i, v := range ProxyPool {
				if v.SuccessNum == 1 {
					if i+1 < len(ProxyPool) {
						ProxyPool = append(ProxyPool[:i], ProxyPool[i+1:]...)
					} else {
						ProxyPool = ProxyPool[:i]
					}
				}
			}
			export()
			lock.Unlock()

		}
	}()
}

func export() {
	//导出代理到文件
	file, err := os.OpenFile("data.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		err.Error()
		return
	}
	data, err := json.Marshal(ProxyPool)
	if err != nil {
		file.Close()
		err.Error()
		return
	}
	file.Write(data)
	file.Close()
}
