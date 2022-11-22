package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

var verifyIS = false
var ProxyPool []ProxyIp
var lock sync.Mutex
var mux2 sync.Mutex

var count int

func countAdd(i int) {
	mux2.Lock()
	count += i
	mux2.Unlock()

}
func countDel() {
	mux2.Lock()
	fmt.Printf("\r代理验证中: %d     ", count)
	count--
	mux2.Unlock()

}
func Verify(pi *ProxyIp, wg *sync.WaitGroup, ch chan int, first bool) {
	defer func() {
		wg.Done()
		countDel()
		<-ch
	}()
	pr := pi.Ip + ":" + pi.Port
	//是抓取验证，还是验证代理池内IP
	startT := time.Now()
	if first {
		if VerifyHttps(pr) {
			pi.Type = "HTTPS"
		} else if VerifyHttp(pr) {
			pi.Type = "HTTP"

		} else if VerifySocket5(pr) {
			pi.Type = "SOCKET5"
		} else {
			return
		}
		tc := time.Since(startT)
		pi.Time = time.Now().Format("2006-01-02 15:04:05")
		pi.Speed = fmt.Sprintf("%s", tc)
		anonymity := Anonymity(pi, 0)
		if anonymity == "" {
			return
		}
		pi.Anonymity = anonymity
	} else {
		pi.RequestNum++
		if pi.Type == "HTTPS" {
			if VerifyHttps(pr) {
				pi.SuccessNum++
			}
		} else if pi.Type == "HTTP" {
			if VerifyHttp(pr) {
				pi.SuccessNum++
			}
		} else if pi.Type == "SOCKET5" {
			if VerifySocket5(pr) {
				pi.SuccessNum++
			}
		}
		tc := time.Since(startT)
		pi.Time = time.Now().Format("2006-01-02 15:04:05")
		pi.Speed = fmt.Sprintf("%s", tc)
		return
	}
	tr := http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Timeout: 15 * time.Second, Transport: &tr}
	//处理返回结果
	res, err := client.Get("https://searchplugin.csdn.net/api/v1/ip/get?ip=" + pi.Ip)
	if err != nil {
		res, err = client.Get("https://searchplugin.csdn.net/api/v1/ip/get?ip=" + pi.Ip)
		if err != nil {
			return
		}
	}
	defer res.Body.Close()
	dataBytes, _ := io.ReadAll(res.Body)
	result := string(dataBytes)
	address := regexp.MustCompile("\"address\":\"(.+?)\",").FindAllStringSubmatch(result, -1)
	if len(address) != 0 {
		addresss := removeDuplication_map(strings.Split(address[0][1], " "))
		le := len(addresss)
		pi.Isp = strings.Split(addresss[le-1], "/")[0]
		for i := range addresss {
			if i == le-1 {
				break
			}
			switch i {
			case 0:
				pi.Country = addresss[0]
			case 1:
				pi.Province = addresss[1]
			case 2:
				pi.City = addresss[2]
			}
		}
	}

	pi.RequestNum = 1
	pi.SuccessNum = 1
	PIAdd(pi)
}
func VerifyHttp(pr string) bool {
	proxyUrl, proxyErr := url.Parse("http://" + pr)
	if proxyErr != nil {
		return false
	}
	tr := http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	tr.Proxy = http.ProxyURL(proxyUrl)
	client := http.Client{Timeout: 10 * time.Second, Transport: &tr}
	request, err := http.NewRequest("GET", "http://baidu.com", nil)
	//处理返回结果
	res, err := client.Do(request)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	dataBytes, _ := io.ReadAll(res.Body)
	result := string(dataBytes)
	if strings.Contains(result, "0;url=http://www.baidu.com") {
		return true
	}
	return false
}
func VerifyHttps(pr string) bool {
	destConn, err := net.DialTimeout("tcp", pr, 10*time.Second)
	if err != nil {
		return false
	}
	defer destConn.Close()
	req := []byte{67, 79, 78, 78, 69, 67, 84, 32, 119, 119, 119, 46, 98, 97, 105, 100, 117, 46, 99, 111, 109, 58, 52, 52, 51, 32, 72, 84, 84, 80, 47, 49, 46, 49, 13, 10, 72, 111, 115, 116, 58, 32, 119, 119, 119, 46, 98, 97, 105, 100, 117, 46, 99, 111, 109, 58, 52, 52, 51, 13, 10, 85, 115, 101, 114, 45, 65, 103, 101, 110, 116, 58, 32, 71, 111, 45, 104, 116, 116, 112, 45, 99, 108, 105, 101, 110, 116, 47, 49, 46, 49, 13, 10, 13, 10}
	destConn.Write(req)
	bytes := make([]byte, 1024)
	destConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	read, err := destConn.Read(bytes)
	if strings.Contains(string(bytes[:read]), "200 Connection established") {
		return true
	}
	return false
}

func VerifySocket5(pr string) bool {
	destConn, err := net.DialTimeout("tcp", pr, 10*time.Second)
	if err != nil {
		return false
	}
	defer destConn.Close()
	req := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	destConn.Write(req)
	bytes := make([]byte, 1024)
	destConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, err = destConn.Read(bytes)
	if err != nil {
		return false
	}
	if bytes[0] == 5 && bytes[1] == 255 {
		return true
	}
	return false

}
func Anonymity(pr *ProxyIp, c int) string {
	c++
	host := "http://httpbin.org/get"
	proxy := ""
	if pr.Type == "SOCKET5" {
		proxy = "socks5://" + pr.Ip + ":" + pr.Port
	} else {
		proxy = "http://" + pr.Ip + ":" + pr.Port
	}
	proxyUrl, proxyErr := url.Parse(proxy)
	if proxyErr != nil {
		if c >= 3 {
			return ""
		}
		return Anonymity(pr, c)
	}
	tr := http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Timeout: 15 * time.Second, Transport: &tr}
	tr.Proxy = http.ProxyURL(proxyUrl)
	request, err := http.NewRequest("GET", host, nil)
	request.Header.Add("Proxy-Connection", "keep-alive")
	//处理返回结果
	res, err := client.Do(request)
	if err != nil {
		if c >= 3 {
			return ""
		}
		return Anonymity(pr, c)
	}
	defer res.Body.Close()
	dataBytes, _ := io.ReadAll(res.Body)
	result := string(dataBytes)
	if !strings.Contains(result, `"url": "http://httpbin.org/`) {
		if c == 3 {
			return ""
		}
		c++
		return Anonymity(pr, c)
	}
	origin := regexp.MustCompile("(\\d+?\\.\\d+?.\\d+?\\.\\d+?,.+\\d+?\\.\\d+?.\\d+?\\.\\d+?)").FindAllStringSubmatch(result, -1)
	if len(origin) != 0 {
		return "透明"
	}
	if strings.Contains(result, "keep-alive") {
		return "普匿"
	}
	return "高匿"
}

func PIAdd(pi *ProxyIp) {
	lock.Lock()
	defer lock.Unlock()
	for i := range ProxyPool {
		if ProxyPool[i].Ip == pi.Ip && ProxyPool[i].Port == pi.Port {
			return
		}
	}
	ProxyPool = append(ProxyPool, *pi)
	ProxyPool = uniquePI(ProxyPool)
}

func VerifyProxy() {
	if run {
		log.Println("代理抓取中, 无法进行代理验证")
		return
	}
	verifyIS = true

	log.Printf("开始验证代理存活情况, 验证次数是当前代理数的4倍: %d\n", len(ProxyPool)*4)
	for i, _ := range ProxyPool {
		ProxyPool[i].RequestNum = 0
		ProxyPool[i].SuccessNum = 0
	}
	count = len(ProxyPool) * 5

	for io := 0; io < 5; io++ {
		for i := range ProxyPool {
			wg3.Add(1)
			ch1 <- 1
			go Verify(&ProxyPool[i], &wg3, ch1, false)
		}
		time.Sleep(15 * time.Second)
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
	log.Printf("\r%s 代理验证结束, 当前可用IP数: %d\n", time.Now().Format("2006-01-02 15:04:05"), len(ProxyPool))
	verifyIS = false
}

func removeDuplication_map(arr []string) []string {
	set := make(map[string]struct{}, len(arr))
	j := 0
	for _, v := range arr {
		_, ok := set[v]
		if ok {
			continue
		}
		set[v] = struct{}{}
		arr[j] = v
		j++
	}

	return arr[:j]
}
