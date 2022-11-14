package main

import (
	"bufio"
	"encoding/json"
	"log"
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
	//设置线程数量
	ch1 = make(chan int, conf.Config.ThreadNum)
	ch2 = make(chan int, conf.Config.ThreadNum)
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
					log.Printf("代理数量不足 %d\n", conf.Config.ProxyNum)
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
		ticker := time.NewTicker(verifyTime * time.Second)
		for range ticker.C {
			log.Println("开始验证代理存活情况")
			for i, _ := range ProxyPool {
				ProxyPool[i].RequestNum = 0
				ProxyPool[i].SuccessNum = 0
			}
			for io := 0; io < 4; io++ {
				for i := range ProxyPool {
					wg3.Add(1)
					ch1 <- 1
					go Verify(&ProxyPool[i], &wg3, ch1, false)
				}
			}
			wg3.Wait()
			lock.Lock()
			var pp []ProxyIp
			for i := range ProxyPool {
				if ProxyPool[i].SuccessNum != 0 {
					pp = append(pp, ProxyPool[i])
				}
			}
			ProxyPool = pp
			export()
			lock.Unlock()
			log.Printf("验证结束,可用IP数: %d\n", len(ProxyPool))
		}

	}()
}

func export() {
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
