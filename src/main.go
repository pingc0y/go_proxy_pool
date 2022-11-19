package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var wg3 sync.WaitGroup
var mux1 sync.Mutex
var ch1 = make(chan int, 50)

func main() {
	//VerifyHttp("127.0.0.1:10809")

	fmt.Println("           ___                     ___            _ " +
		"\n ___  ___ | . " + "\\ _ _  ___ __   _ _ | . \\ ___  ___ | |" +
		"\n/ . |/ . \\|  _/| '_>/ . \\\\ \\/| | ||  _// . \\/ . \\| |" +
		"\n\\_. |\\___/|_|  |_|  \\___//\\_\\`_. ||_|  \\___/\\___/|_|" +
		"\n<___'                        <___'                  ")
	InitData()
	//开启隧道代理
	go httpSRunTunnelProxyServer()
	go socket5RunTunnelProxyServer()
	//启动webAPi
	Run()
}

// 初始化
func InitData() {
	//获取配置文件
	GetConfigData()
	//设置线程数量
	ch1 = make(chan int, conf.Config.ThreadNum)
	ch2 = make(chan int, conf.Config.ThreadNum)
	//是否需要抓代理
	if len(ProxyPool) < conf.Config.ProxyNum {
		//抓取代理
		spiderRun()
	}
	//定时判断是否需要获取代理iP
	go func() {
		// 每 60 秒钟时执行一次
		ticker := time.NewTicker(60 * time.Second)
		for range ticker.C {
			if len(ProxyPool) < conf.Config.ProxyNum {
				if !run {
					log.Printf("代理数量不足 %d\n", conf.Config.ProxyNum)
					//抓取代理
					spiderRun()
				}
			} else {
				//保存代理到本地
				export()
			}
		}
	}()

	//定时更换隧道IP
	go func() {
		tunnelTime := time.Duration(conf.Config.TunnelTime)
		ticker := time.NewTicker(tunnelTime * time.Second)
		for range ticker.C {
			if len(ProxyPool) != 0 {
				httpsIp = getHttpsIp()
				httpIp = gethttpIp()
				socket5Ip = getSocket5Ip()
			}
		}
	}()

	// 验证代理存活情况
	go func() {
		verifyTime := time.Duration(conf.Config.VerifyTime)
		ticker := time.NewTicker(verifyTime * time.Second)
		for range ticker.C {
			if !verifyIS {
				VerifyProxy()
			}
		}
	}()
}

func export() {
	mux1.Lock()
	defer mux1.Unlock()
	//导出代理到文件
	err := os.Truncate("data.json", 0)
	if len(ProxyPool) == 0 {
		return
	}
	if err != nil {
		log.Printf("data.json清理失败：%s", err)
		return
	}
	file, err := os.OpenFile("data.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("data.json打开失败：%s", err)
		return
	}
	defer file.Close()

	data, err := json.Marshal(ProxyPool)
	if err != nil {
		log.Printf("代理json化失败：%s", err)
		return
	}
	buf := bufio.NewWriter(file)
	// 字节写入
	buf.Write(data)
	// 将缓冲中的数据写入
	err = buf.Flush()
	if err != nil {
		log.Println("代理json保存失败:", err)
	}
}
