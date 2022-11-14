package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
)

type Home struct {
	Sum         int               `yaml:"sum" json:"sum"`
	TunnelProxy map[string]string `yaml:"tunnelProxy" json:"tunnelProxy"`
	Type        map[string]int    `yaml:"type" json:"type"`
	Anonymity   map[string]int    `yaml:"anonymity" json:"anonymity"`
	Region      map[string]int    `yaml:"region" json:"region"`
	Source      map[string]int    `yaml:"source" json:"source"`
}

var record []ProxyIp

func Run() {
	log.Printf(" webApi启动 - 监听IP端口 -> %s\n", conf.Config.Ip+":"+conf.Config.Port)
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	//首页
	r.GET("/", func(c *gin.Context) {
		home := Home{Sum: len(ProxyPool), Type: make(map[string]int), Anonymity: make(map[string]int), Region: make(map[string]int), Source: make(map[string]int), TunnelProxy: make(map[string]string)}
		for i := range ProxyPool {
			home.Type[ProxyPool[i].Type] += 1
			home.Anonymity[ProxyPool[i].Anonymity] += 1
			home.Region[ProxyPool[i].Info1] += 1
			home.Source[ProxyPool[i].Source] += 1
		}
		home.TunnelProxy["HTTP"] = Iip
		home.TunnelProxy["HTTPS"] = Sip
		jsonByte, _ := json.Marshal(&home)
		jsonStr := string(jsonByte)
		c.String(200, jsonStr)
	})

	//查询
	r.GET("/get", func(c *gin.Context) {
		if len(ProxyPool) == 0 {
			c.String(200, fmt.Sprintf("{\"code\": 200, \"msg\": \"代理池是空的\"}"))
			return
		}
		var prs []ProxyIp
		var jsonByte []byte
		ty := c.DefaultQuery("type", "all")
		an := c.DefaultQuery("anonymity", "all")
		re := c.DefaultQuery("region", "all")
		so := c.DefaultQuery("source", "all")
		co := c.DefaultQuery("count", "1")
		for _, v := range ProxyPool {
			if (v.Type == ty || ty == "all") && (v.Anonymity == an || an == "all") && (v.Info1 == re || re == "all") && (v.Source == so || so == "all") {
				prs = append(prs, v)
			}
		}
		if co == "all" {
			jsonByte, _ = json.Marshal(prs)
		} else if co == "1" {
			var _is bool
			for _, v := range prs {
				_is = true
				for _, vv := range record {
					if v.Ip+v.Port == vv.Ip+vv.Port {
						_is = false
					}
				}
				if _is {
					jsonByte, _ = json.Marshal(v)
					record = append(record, v)
					break
				}
			}
			if !_is {
				jsonByte, _ = json.Marshal(prs[0])
				record = []ProxyIp{prs[0]}
			}
		} else {
			count, err := strconv.Atoi(co)
			if err != nil {
				c.String(500, fmt.Sprintf("{\"code\": 500, \"msg\": \"错误\"}"))
			}
			jsonByte, _ = json.Marshal(prs[:count])
		}
		jsonStr := string(jsonByte)
		c.String(200, jsonStr)
	})

	//删除
	r.GET("/del", func(c *gin.Context) {
		if len(ProxyPool) == 0 {
			c.String(200, fmt.Sprintf("{\"code\": 200, \"msg\": \"代理池是空的\"}"))
			return
		}
		ip := c.Query("ip")
		port := c.Query("port")
		i := delIp(ip + ":" + port)
		c.String(200, fmt.Sprintf("{\"code\": 200, \"count\": %d}", i))
	})
	//抓取代理
	r.GET("/upload", func(c *gin.Context) {
		if run {
			c.String(200, fmt.Sprintf("{\"code\": 200, \"msg\": \"抓取中\"}"))
		} else {
			spiderRun()
			export()
			c.String(200, fmt.Sprintf("{\"code\": 200, \"msg\": \"开始抓取代理IP\"}"))
		}
	})

	//更换隧道代理IP
	r.GET("/tunnelUpdate", func(c *gin.Context) {
		if len(ProxyPool) == 0 {
			c.String(200, fmt.Sprintf("{\"code\": 200, \"msg\": \"代理池是空的\"}"))
		}
		Iip = getIIp()
		Sip = getSIp()
		c.String(200, fmt.Sprintf("{\"code\": 200, \"HTTP\": \"%s\",\"HTTPS\": \"%s\"}", Iip, Sip))
	})

	r.Run(conf.Config.Ip + ":" + conf.Config.Port)
}

func delIp(addr string) int {
	lock.Lock()
	defer lock.Unlock()
	var in int
	for i, v := range ProxyPool {
		if v.Ip+":"+v.Port == addr {
			in++
			if i+1 < len(ProxyPool) {
				ProxyPool = append(ProxyPool[:i], ProxyPool[i+1:]...)
			} else {
				ProxyPool = ProxyPool[:i]
			}
		}
	}
	return in

}
