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
	v := flag.Bool("v", false, "version update log.")
	p := flag.String("p", "8888", "port, 8888 as default.")
	l := flag.String("l", "is.log", "log file path, ./is.log as default.")
	c := flag.String("c", "conf.toml", "plugins path, ./scripts as default.")
	s := flag.String("s", "scripts/", "plugins path, ./scripts/ as default, must be end with '/'.")
	r := flag.String("r", "html/", "response body file path, ./html/ as default, must be end with '/'.")
	g := flag.Int("g", web.DefaultGzip, "response body gzip type, DefaultGzip as default. It can be BestGzip, DefaultGzip or NoGzip.")
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

	web.KAccess.Thread()
	go web.KAccess.Start()

	web.WebServer(port, *g)
}
