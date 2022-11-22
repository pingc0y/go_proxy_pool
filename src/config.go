package main

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"os"
)

var conf *Config

type Config struct {
	Spider       []Spider       `yaml:"spider" json:"spider"`
	SpiderPlugin []SpiderPlugin `yaml:"spiderPlugin" json:"spiderPlugin"`
	SpiderFile   []SpiderFile   `yaml:"spiderFile" json:"spiderFile"`
	Proxy        Proxy          `yaml:"proxy" json:"proxy"`
	Config       config         `yaml:"config" json:"config"`
}
type config struct {
	Ip               string `yaml:"ip" json:"ip"`
	Port             string `yaml:"port" json:"port"`
	HttpTunnelPort   string `yaml:"httpTunnelPort" json:"httpTunnelPort"`
	SocketTunnelPort string `yaml:"socketTunnelPort" json:"socketTunnelPort"`
	TunnelTime       int    `yaml:"tunnelTime" json:"tunnelTime"`
	ProxyNum         int    `yaml:"proxyNum" json:"proxyNum"`
	VerifyTime       int    `yaml:"verifyTime" json:"verifyTime"`
	ThreadNum        int    `yaml:"threadNum" json:"threadNum"`
}
type Spider struct {
	Name     string            `yaml:"name" json:"name"`
	Method   string            `yaml:"method" json:"method"`
	Interval int               `yaml:"interval" json:"interval"`
	Body     string            `yaml:"body" json:"body"`
	ProxyIs  bool              `yaml:"proxy" json:"proxy"`
	Headers  map[string]string `yaml:"headers" json:"headers"`
	Urls     string            `yaml:"urls" json:"urls"`
	Ip       string            `yaml:"ip" json:"ip"`
	Port     string            `yaml:"port" json:"port"`
}
type SpiderPlugin struct {
	Name string `yaml:"name" json:"name"`
	Run  string `yaml:"run" json:"run"`
}
type SpiderFile struct {
	Name string `yaml:"name" json:"name"`
	Path string `yaml:"path" json:"path"`
}

type Proxy struct {
	Host string `yaml:"host" json:"host"`
	Port string `yaml:"port" json:"port"`
}
type ProxyIp struct {
	Ip         string //IP地址
	Port       string //代理端口
	Country    string //代理国家
	Province   string //代理省份
	City       string //代理城市
	Isp        string //IP提供商
	Type       string //代理类型
	Anonymity  string //代理匿名度, 透明：显示真实IP, 普匿：显示假的IP, 高匿：无代理IP特征
	Time       string //代理验证
	Speed      string //代理响应速度
	SuccessNum int    //验证请求成功的次数
	RequestNum int    //验证请求的次数
	Source     string //代理源
}

// 数组去重
func uniquePI(arr []ProxyIp) []ProxyIp {
	var pr []ProxyIp
	for _, v := range arr {
		is := true
		for _, vv := range pr {
			if v.Ip+v.Port == vv.Ip+vv.Port {
				is = false
			}
		}
		if is {
			pr = append(pr, v)
		}
	}
	return pr
}

// 读取配置文件
func GetConfigData() {
	//导入配置文件
	yamlFile, err := os.ReadFile("config.yml")
	if err != nil {
		log.Println("配置文件打开错误：" + err.Error())
		err.Error()
		return
	}
	//将配置文件读取到结构体中
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		log.Println("配置文件解析错误：" + err.Error())
		err.Error()
		return
	}
	//导入代理缓存
	file, err := os.OpenFile("data.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println("代理json文件打开错误：" + err.Error())
		err.Error()
		return
	}
	defer file.Close()
	all, err := io.ReadAll(file)
	if err != nil {
		log.Println("代理json解析错误：" + err.Error())
		return
	}
	if len(all) == 0 {
		return
	}
	err = json.Unmarshal(all, &ProxyPool)
	if err != nil {
		log.Println("代理json解析错误：" + err.Error())
		return
	}

}

// 处理Headers配置
func SetHeadersConfig(he map[string]string, header *http.Header) *http.Header {
	for k, v := range he {
		header.Add(k, v)
	}
	return header
}
