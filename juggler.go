package main

import (
	"Juggler/config"
	"Juggler/golua"
	"Juggler/logger"
	"Juggler/web"
	"flag"
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"strconv"
)

var versionInfo = `[*] 更新说明
[+] 2020年5月14日
    增加geoip数据
    允许通过-p参数进行端口指定
    允许通过-l参数进行日志文件路径指定
    允许通过-c参数进行配置文件路径指定

[+] 2020年6月20日
    把toml配置文件替换成lua插件模式

[+] 2020年6月25日
    增加time库和crypto库
    使用kafka做消息缓冲入ES,使用gzip压缩减少网络io
    通过lua虚拟机进行web请求和响应的基本处理

[+] 2020年6月30日
    可以通过lua插件自定义返回内容,响应格式包括
        1.只状态码
        2.固定格式化字符串+状态码
        3.格式化文件内容+状态码
        4.克隆站内容+状态码
    响应体内容缓存进内存减少文件读取io

[+] 2020年7月1日
    响应体使用gzip减少网络io
    将原有的拦截中心toml配置全部转化成lua脚本
`

func main() {

	var err error
	v := flag.Bool("v", false, "版本参数")
	p := flag.String("p", "8888", "端口参数 默认端口8888")
	l := flag.String("l", "is.log", "日志路径参数 默认是./is.log")
	c := flag.String("c", "conf.toml", "配置文件路径参数 默认是./conf.toml")
	s := flag.String("s", "scripts/", "插件路径参数 默认是./scripts/")
	r := flag.String("r", "html/", "响应内容文件路径参数 默认是./html/ 且自定义路径必须以/结尾")
	g := flag.Int("g", web.DefaultGzip, "响应内容gzip压缩参数 默认DefaultGzip 0是不压缩 1是默认压缩 2是极致压缩")
	k := flag.Int("k", web.KafkaOpen, "kafka线程启动开关参数 默认开 0是开 其他都是关")
	// 初始化应用参数
	flag.Parse()
	// 初始化日志打印器
	if *l == "" {
		err = logger.NewPrinter("is.log")
	} else {
		err = logger.NewPrinter(*l)
	}
	if err != nil {
		fmt.Printf("new printer error : %v", err)
		return
	}
	// 初始化配置，现在只有kafka配置
	config.Cfg = config.Config{}
	var cp string
	if *c == "" {
		cp = "conf.toml"
	} else {
		cp = *c
	}
	if config.Cfg.Load(cp) != nil {
		logger.Printer(logger.ERROR, "config init error.")
	}
	go config.Cfg.Monitor(cp)
	// 打印当前版本信息
	if *v {
		fmt.Println(versionInfo)
		return
	}
	// 初始化监听端口，默认8888
	var port string
	if *p == "" {
		port = ":8888"
	} else {
		_, err := strconv.Atoi(*p)
		if err != nil {
			logger.Printer(logger.ERROR, "bad input : %v", err)
			port = ":8888"
		} else {
			port = ":" + *p
		}
	}
	// 初始化插件路径
	golua.LuaPool = &golua.LStatePool{
		VMs:   make(chan *lua.LState, 10240),
		Htmls: make(map[string][]byte),
		Fns:   make(map[string]*lua.LFunction),
	}
	if err = golua.LuaPool.Init(*s); err != nil {
		return
	}
	// 初始化相应内容表
	if err = golua.LuaPool.LoadHtml(*r); err != nil {
		logger.Printer(logger.ERROR, "load htmls in %s error : %v", *r, err)
		return
	}
	defer golua.LuaPool.Shutdown()

	if *k == 0 {
		web.KAccess.Thread()
		go web.KAccess.Start()
	}

	web.WebServer(port, *g)
}
