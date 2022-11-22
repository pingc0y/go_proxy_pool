# goProxyPool

一款无环境依赖开箱即用的免费代理IP池   

内置14个免费代理源，均使用内置的简单正则获取  

支持调用插件扩展代理源，返回的数据符合格式即可，无开发语言限制  

支持webApi获取、删除、更新等代理池内的IP 

支持 http，socket5 隧道代理模式，无需手动更换IP  

遇到bug或有好的建议，欢迎提issue

## 隧道代理
隧道代理是代理IP存在的一种方式。  
相对于传统固定代理IP，它的特点是自动地在代理服务器上改变IP，这样每个请求都使用一个不同的IP。

## 代理IP特征
这里提供一些代理IP的特征，师傅们可通过特征自己写代理源，api获取的话内置的正则方式就能写  
360网络空间测绘_socket5：
```text
protocol:"socks5" AND "Accepted Auth Method: 0x0" AND "connection: close" AND country: "China"  
```
fofa_http:
```text  
"HTTP/1.1 403 Forbidden Server: nginx/1.12.1" && port="9091"   

port="3128" && title="ERROR: The requested URL could not be retrieved"  

"X-Cache: 'MISS from VideoCacheBox/CE8265A63696DECD7F0D17858B1BDADC37771805'" && "X-Squid-Error: ERR_ACCESS_DENIED 0"  
```
hunter_http：
```text
header.server="nginx/2.2.200603d"&&web.title="502 Bad Gateway" && ip.port="8085"
```

# 截图
[![zuz6TU.png](https://s1.ax1x.com/2022/11/19/zuz6TU.png)](https://s1.ax1x.com/2022/11/19/zuz6TU.png)
# 使用说明
下载 
```
git clone https://github.com/pingc0y/go_proxy_pool.git
```
编译（直接使用成品，就无需编译）  
以下是在windows环境下，编译出各平台可执行文件的命令  
```
SET CGO_ENABLED=0
SET GOOS=windows
SET GOARCH=amd64
go build -ldflags "-s -w" -o ../goProxyPool-windows-amd64.exe

SET CGO_ENABLED=0
SET GOOS=windows
SET GOARCH=386
go build -ldflags "-s -w"  -o ../goProxyPool-windows-386.exe

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -ldflags "-s -w" -o ../goProxyPool-linux-amd64

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=arm64
go build -ldflags "-s -w" -o ../goProxyPool-linux-arm64

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=386
go build -ldflags "-s -w" -o ../goProxyPool-linux-386

SET CGO_ENABLED=0
SET GOOS=darwin
SET GOARCH=amd64
go build -ldflags "-s -w" -o ../goProxyPool-macos-amd64

SET CGO_ENABLED=0
SET GOOS=darwin
SET GOARCH=arm64
go build -ldflags "-s -w" -o ../goProxyPool-macos-arm64

```
运行  
需要与config.yml在同一目录  
注意：抓取代理会进行类型地区等验证会比较缓慢，存活验证会快很多
```
.\goProxyPool.exe
```

代理源中有部分需要翻墙才能访问，有条件就设置下config.yml的代理配置
```yml
proxy:
  host: 127.0.0.1
  port: 10809
```
## webAPi说明
查看代理池情况
```
http://127.0.0.1:8080/
```

获取代理
```
http://127.0.0.1:8080/get?type=HTTP&count=10&anonymity=all
可选参数：
type        代理类型
anonymity   匿名度
country     国家
source      代理源
count       代理数量
获取所有：all
```

删除代理
```
http://127.0.0.1:8080/delete?ip=127.0.0.1&port=8888
必须传参：
ip      代理ip
port    代理端口
```

验证代理
```
http://127.0.0.1:8080/verify
```

更换隧道代理IP
```
http://127.0.0.1:8080/tunnelUpdate
```

抓取代理
```
http://127.0.0.1:8080/spider
```
## 代理字段解读
```go
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
```
## 配置文件
```yaml
#使用代理去获取代理IP
proxy:
  host: 127.0.0.1
  port: 10809

# 配置信息
config:
  #监听IP
  ip: 0.0.0.0
  #web监听端口
  port: 8080
  #http隧道代理端口
  httpTunnelPort: 8111
  #socket隧道代理端口
  socketTunnelPort: 8112
  #隧道代理更换时间秒
  tunnelTime: 60
  #可用IP数量小于‘proxyNum’时就去抓取
  proxyNum: 50
  #代理IP验证间隔秒
  verifyTime: 1800
  #抓取/检测状态线程数
  threadNum: 200

#ip源
spider:
    #代理获取源1
  - name: '齐云代理'
    #请求方式
    method: 'GET'
    #POST传参用的请求体
    body: ''
    #urls请求间隔/秒，防止频率过快被限制
    interval: 0
    #使用的请求头
    Headers:
      User-Agent: 'Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)'
    #获取的地址
    urls: 'https://proxy.ip3366.net/free/?action=china&page=1,https://proxy.ip3366.net/free/?action=china&page=2,https://proxy.ip3366.net/free/?action=china&page=3'
    #获取IP的正则表达式，
    ip: '\"IP\">(\d+?\.\d+?.\d+?\.\d+?)</td>'
    #获取端口的正则表达式
    port: '\"PORT\">(\d+?)</td>'
    #是否使用代理去请求
    proxy: false
      
#通过插件，扩展ip源
spiderPlugin:
  #插件名
  - name: test
    #运行命令，返回的结果要符合格式
    run: '.\test1.exe'
    
#通过文件，导入IP
spiderFile:
  #文件名
  - name: test1
    #文件路径
    path: 'ip.txt'
```
### 扩展返回格式
通过,分割
```text
110.179.64.89:1080,111.2.155.180:1090,111.172.3.212:1090,111.196.186.95:6669
```
### 文件导入格式
通过换行分割
```text
110.179.64.89:1080
111.2.155.180:1090
111.172.3.212:1090
111.196.186.95:6669
111.201.103.29:1080
113.12.200.66:1080
113.67.96.67:1090
113.104.217.45:1080
113.110.246.76:1080
113.116.9.18:1080
113.119.193.183:1090
113.119.193.187:1090
113.249.93.219:1080
114.95.200.164:1080
115.193.161.177:1080
```


## 更新说明
2022/11/22  
修复 ip归属地接口更换  
优化 验证代理   

2022/11/19  
新增 socket5代理  
新增 文件导入代理  
新增 显示验证进度  
新增 验证webApi  
修改 扩展导入格式  
优化 代理验证方式  
优化 匿名度改为自动识别  
修复 若干bug  


