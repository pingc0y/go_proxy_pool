package main

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

var wg sync.WaitGroup
var wg2 sync.WaitGroup
var mux sync.Mutex
var ch2 = make(chan int, 50)

var run = false

func spiderRun() {
	run = true
	defer func() {
		run = false
	}()
	log.Println("开始抓取代理...")
	for i := range conf.Spider {
		wg2.Add(1)
		go spider(&conf.Spider[i])
	}
	wg2.Wait()

	log.Println("抓取結束")
	log.Println("开始扩展抓取代理...")
	for i := range conf.SpiderPlugin {
		wg2.Add(1)
		go spiderPlugin(&conf.SpiderPlugin[i])
	}
	wg2.Wait()
	log.Println("扩展抓取結束")
	run = false

}

func spider(sp *Spider) {
	defer func() {
		wg2.Done()
		//log.Printf("%s 结束...",sp.Name)
	}()
	//log.Printf("%s 开始...", sp.Name)
	urls := strings.Split(sp.Urls, ",")
	var pis []ProxyIp
	for _, v := range urls {
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		if sp.ProxyIs {
			proxyUrl, parseErr := url.Parse("http://" + conf.Proxy.Host + ":" + conf.Proxy.Port)
			if parseErr != nil {
				log.Println("代理地址错误: \n" + parseErr.Error())
				return
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
			return
		}
		dataBytes, _ := io.ReadAll(res.Body)
		result := string(dataBytes)
		ip := regexp.MustCompile(sp.Ip).FindAllStringSubmatch(result, -1)
		port := regexp.MustCompile(sp.Port).FindAllStringSubmatch(result, -1)
		anonymity := regexp.MustCompile(sp.Anonymity).FindAllStringSubmatch(result, -1)
		if len(ip) == 0 {
			return
		}
		for i := range ip {
			var _anonymity string
			var _ip string
			var _port string
			if !strings.Contains(sp.Anonymity, "(") && !strings.Contains(sp.Anonymity, ")") {
				_anonymity = sp.Anonymity
			} else {
				if len(anonymity) > i {
					_anonymity, err = url.QueryUnescape(anonymity[i][1])
				} else if len(anonymity) != 0 {
					_anonymity, err = url.QueryUnescape(anonymity[i-1][1])
				} else {
					_anonymity = "透明"
				}
			}
			_ip, _ = url.QueryUnescape(ip[i][1])
			_port, _ = url.QueryUnescape(port[i][1])

			var keys []string
			for key := range sp.Replace {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				_anonymity = strings.Replace(_anonymity, key, sp.Replace[key], -1)
			}

			pis = append(pis, ProxyIp{Ip: _ip, Port: _port, Anonymity: _anonymity, Source: sp.Name})
		}

	}
	pis = uniquePI(pis)
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

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			status := exitErr.Sys().(syscall.WaitStatus)
			switch {
			case status.Exited():
				log.Printf("%s 失败\n", spp.Name)
			case status.Signaled():
				log.Printf("%s 失败\n", spp.Name)
			}
		} else {
			log.Printf("%s 失败\n", spp.Name)
		}
	} else {
		var pis []ProxyIp
		err = json.Unmarshal(buf, &pis)
		if err != nil {
			log.Printf("%s 返回值不符合规范\n", spp.Name)
			return
		}
		pis = uniquePI(pis)
		for i := range pis {
			wg.Add(1)
			ch2 <- 1
			go Verify(&pis[i], &wg, ch2, true)
		}
		wg.Wait()
	}
}
