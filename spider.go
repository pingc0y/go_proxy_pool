package main

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup
var wg2 sync.WaitGroup
var mux sync.Mutex
var ch2 = make(chan int, 50)

// 是否抓取中
var run = false

func spiderRun() {

	run = true
	defer func() {
		run = false
	}()

	count = 0
	log.Println("开始抓取代理...")
	for i := range conf.Spider {
		wg2.Add(1)
		go spider(&conf.Spider[i])
	}
	wg2.Wait()
	log.Printf("\r%s 代理抓取结束           \n", time.Now().Format("2006-01-02 15:04:05"))

	count = 0
	log.Println("开始扩展抓取代理...")
	for i := range conf.SpiderPlugin {
		wg2.Add(1)
		go spiderPlugin(&conf.SpiderPlugin[i])
	}
	wg2.Wait()
	log.Printf("\r%s 扩展代理抓取结束         \n", time.Now().Format("2006-01-02 15:04:05"))
	count = 0
	log.Println("开始文件抓取代理...")
	for i := range conf.SpiderFile {
		wg2.Add(1)
		go spiderFile(&conf.SpiderFile[i])
	}
	wg2.Wait()
	log.Printf("\r%s 文件代理抓取结束         \n", time.Now().Format("2006-01-02 15:04:05"))

	//导出代理到文件
	export()

}

func spider(sp *Spider) {
	defer func() {
		wg2.Done()
		//log.Printf("%s 结束...",sp.Name)
	}()
	//log.Printf("%s 开始...", sp.Name)
	urls := strings.Split(sp.Urls, ",")
	var pis []ProxyIp
	for ui, v := range urls {
		if ui != 0 {
			time.Sleep(time.Duration(sp.Interval) * time.Second)
		}
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		if sp.ProxyIs {
			proxyUrl, parseErr := url.Parse("http://" + conf.Proxy.Host + ":" + conf.Proxy.Port)
			if parseErr != nil {
				log.Println("代理地址错误: \n" + parseErr.Error())
				continue
			}
			tr.Proxy = http.ProxyURL(proxyUrl)
		}
		client := http.Client{Timeout: 20 * time.Second, Transport: tr}
		request, _ := http.NewRequest(sp.Method, v, strings.NewReader(sp.Body))
		//设置请求头
		SetHeadersConfig(sp.Headers, &request.Header)
		//处理返回结果
		res, err := client.Do(request)
		if err != nil {
			continue
		}
		dataBytes, _ := io.ReadAll(res.Body)
		result := string(dataBytes)
		ip := regexp.MustCompile(sp.Ip).FindAllStringSubmatch(result, -1)
		port := regexp.MustCompile(sp.Port).FindAllStringSubmatch(result, -1)
		if len(ip) == 0 {
			continue
		}
		for i := range ip {
			var _ip string
			var _port string
			_ip, _ = url.QueryUnescape(ip[i][1])
			_port, _ = url.QueryUnescape(port[i][1])
			_is := true
			for pi := range ProxyPool {
				if ProxyPool[pi].Ip == _ip && ProxyPool[pi].Port == _port {
					_is = false
					break
				}
			}
			if _is {
				pis = append(pis, ProxyIp{Ip: _ip, Port: _port, Source: sp.Name})
			}
		}
	}
	pis = uniquePI(pis)
	countAdd(len(pis))
	for i := range pis {
		wg.Add(1)
		ch2 <- 1
		go Verify(&pis[i], &wg, ch2, true)
	}
	wg.Wait()

}

func spiderPlugin(spp *SpiderPlugin) {
	defer func() {
		wg2.Done()
	}()
	cmd := exec.Command("cmd.exe", "/c", spp.Run)
	//Start执行不会等待命令完成，Run会阻塞等待命令完成。
	//err := cmd.Start()
	//err := cmd.Run()
	//cmd.Output()函数的功能是运行命令并返回其标准输出。
	buf, err := cmd.Output()
	var pis []ProxyIp
	if err != nil {
		log.Println("失败", spp.Name, err)
	} else {
		_is := true
		line := strings.Split(string(buf), ",")
		for i := range line {
			split := strings.Split(line[i], ":")
			for pi := range ProxyPool {
				if ProxyPool[pi].Ip == split[0] && ProxyPool[pi].Port == split[1] {
					_is = false
					break
				}
			}
			if _is {
				pis = append(pis, ProxyIp{Ip: split[0], Port: split[1], Source: spp.Name})
			}
		}
		//var _pis []ProxyIp
		//var pis []ProxyIp
		//var _is = true
		//err = json.Unmarshal(buf, &_pis)
		//if err != nil {
		//	log.Printf("%s 返回值不符合规范\n", spp.Name)
		//	return
		//}
		//for i := range _pis {
		//	for pi := range ProxyPool {
		//		if ProxyPool[pi].Ip == _pis[i].Ip && ProxyPool[pi].Port == _pis[i].Port {
		//			_is = false
		//			break
		//		}
		//	}
		//	if _is {
		//		pis = append(pis, ProxyIp{Ip: _pis[i].Ip, Port: _pis[i].Port, Source: spp.Name})
		//	}
	}
	pis = uniquePI(pis)
	countAdd(len(pis))
	for i := range pis {
		wg.Add(1)
		ch2 <- 1
		go Verify(&pis[i], &wg, ch2, true)
	}
	wg.Wait()
}

func spiderFile(spp *SpiderFile) {
	defer func() {
		wg2.Done()
	}()
	var pis []ProxyIp
	fi, err := os.Open(spp.Path)
	if err != nil {
		log.Println(spp.Name, "失败", err)
		return
	}
	r := bufio.NewReader(fi) // 创建 Reader
	for {
		_is := true
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			split := strings.Split(strings.TrimSpace(string(line)), ":")
			for pi := range ProxyPool {
				if ProxyPool[pi].Ip == split[0] && ProxyPool[pi].Port == split[1] {
					_is = false
					break
				}
			}
			if _is {
				pis = append(pis, ProxyIp{Ip: split[0], Port: split[1], Source: spp.Name})
			}
		}
		if err != nil {
			break
		}
	}
	pis = uniquePI(pis)
	countAdd(len(pis))
	for i := range pis {
		wg.Add(1)
		ch2 <- 1
		go Verify(&pis[i], &wg, ch2, true)
	}
	wg.Wait()

}
