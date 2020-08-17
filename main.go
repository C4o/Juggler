package main

import (
	"IUS/conf"
	"IUS/web"
	"flag"
	"fmt"
	"log"
	"strconv"
)

// short name for Intercept Server.

var agentInfo = `[*] 更新说明

[+] 2020年5月14日更新
    增加geoip数据
    允许通过-p参数进行端口指定
    允许通过-l参数进行日志文件路径指定
    允许通过-c参数进行配置文件路径指定
`

func main() {

	v := flag.Bool("v", false, "agent info")
	p := flag.String("p", "8888", "端口号,默认8888")
	l := flag.String("l", "is.log", "")
	c := flag.String("c", "conf.toml", "")

	flag.Parse()

	if *c == "" {
		conf.ConfigFile = "conf.toml"
	} else {
		conf.ConfigFile = *c
	}

	if *l == "" {
		web.LogFile = "is.log"
	} else {
		web.LogFile = *l
	}

	if *v {
		fmt.Println(agentInfo)
		return
	}

	var port string
	if *p == "" {
		port = ":8888"
	} else {
		_, err := strconv.Atoi(*p)
		if err != nil {
			log.Printf("bad input : %v", err)
			port = ":8888"
		} else {
			port = ":" + *p
		}
	}

	conf.Cfg = conf.Config{}
	conf.Cfg.Load()
	web.Kafka.Thread()
	if err := web.NewPrinter(); err != nil {
		fmt.Printf("new printer error : %v", err)
		return
	}
	go web.Kafka.Start()
	go web.WebServer(port)
	conf.Cfg.Monitor()
}
