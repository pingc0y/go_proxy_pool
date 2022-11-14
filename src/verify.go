package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

type ProxyIp struct {
	Ip         string
	Port       string
	Info1      string
	Info2      string
	Info3      string
	Isp        string
	Type       string
	Anonymity  string
	Time       int64
	Speed      int64
	SuccessNum int
	RequestNum int
	Source     string
}

var ProxyPool []ProxyIp
var lock sync.Mutex

func Verify(pi *ProxyIp, wg *sync.WaitGroup, ch chan int) {
	defer func() {
		wg.Done()
		<-ch
	}()
	startT := time.Now()
	//配置代理
	proxyUrl, parseErr := url.Parse("http://" + pi.Ip + ":" + pi.Port)
	if parseErr != nil {
		return
	}
	tr := http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	tr.Proxy = http.ProxyURL(proxyUrl)
	client := http.Client{Timeout: 20 * time.Second, Transport: &tr}
	host := "http://api.vore.top/api/IPv4"
	request, _ := http.NewRequest("GET", host, nil)
	//处理返回结果
	res, err := client.Do(request)
	pi.RequestNum += 1
	if err != nil {
		return
	}
	dataBytes, _ := io.ReadAll(res.Body)
	result := string(dataBytes)
	if !strings.Contains(result, "info1") {
		return
	}
	info1 := regexp.MustCompile("info1\": \"(.*?)\"").FindAllStringSubmatch(result, -1)
	if len(info1) != 0 {
		pi.Info1 = info1[0][1]
		info2 := regexp.MustCompile("info2\": \"(.*?)\"").FindAllStringSubmatch(result, -1)
		if len(info2) != 0 {
			pi.Info2 = info2[0][1]
			info3 := regexp.MustCompile("info3\": \"(.*?)\"").FindAllStringSubmatch(result, -1)
			if len(info3) != 0 {
				pi.Info3 = info3[0][1]
			}
		}
	}
	isp := regexp.MustCompile("isp\": \"(.*?)\"").FindAllStringSubmatch(result, -1)
	if len(isp) != 0 {
		pi.Isp = isp[0][1]
	}
	tc := time.Since(startT)
	pi.Time = time.Now().Unix()
	pi.Speed = int64(tc)
	pi.SuccessNum++
	if pi.RequestNum == 1 {
		if HTTPSVerify(pi.Ip + ":" + pi.Port) {
			pi.Type = "HTTPS"
		} else {
			pi.Type = "HTTP"
		}
		PIAdd(pi)
	}
}

func PIAdd(pi *ProxyIp) {
	lock.Lock()
	defer lock.Unlock()
	ProxyPool = append(ProxyPool, *pi)
	ProxyPool = uniquePI(ProxyPool)
}

func HTTPSVerify(pr string) bool {
	destConn, err := net.DialTimeout("tcp", pr, 10*time.Second)
	if err != nil {
		return false
	}
	req := []byte{67, 79, 78, 78, 69, 67, 84, 32, 119, 119, 119, 46, 98, 97, 105, 100, 117, 46, 99, 111, 109, 58, 52, 52, 51, 32, 72, 84, 84, 80, 47, 49, 46, 49, 13, 10, 72, 111, 115, 116, 58, 32, 119, 119, 119, 46, 98, 97, 105, 100, 117, 46, 99, 111, 109, 58, 52, 52, 51, 13, 10, 85, 115, 101, 114, 45, 65, 103, 101, 110, 116, 58, 32, 71, 111, 45, 104, 116, 116, 112, 45, 99, 108, 105, 101, 110, 116, 47, 49, 46, 49, 13, 10, 13, 10}
	destConn.Write(req)
	bytes := make([]byte, 1024)
	destConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	read, err := destConn.Read(bytes)
	if strings.Contains(fmt.Sprintf("%s", bytes[:read]), "200 Connection established") {
		return true
	}
	return false
}
