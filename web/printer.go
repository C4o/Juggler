package web

import (
	"IUS/conf"
	"fmt"
	"log"
	"os"
)

var (
	ERROR   = 0
	INFO    = 2
	DEBUG   = 4
	LogFile string
)

func NewPrinter() error {

	output, err := os.OpenFile(LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	log.SetOutput(output)

	return nil
}

func Printer(level int, format string, v ...interface{}) {

	debug := conf.Cfg.Default.Debug

	// 如果等级一样，就只打印该等级而没有其他等级日志
	if level == debug {
		logging(level, format, v...)
	} else {
		// 如果等级不一样，就把小于配置文件等级的所有日志等级全部打印
		if debug > level {
			logging(level, format, v...)
		}
	}
}

func logging(level int, format string, v ...interface{}) {

	switch level {
	case 0:
		log.Println(fmt.Sprintf("[error] "+format, v...))
	case 2:
		log.Println(fmt.Sprintf("[info] "+format, v...))
	case 4:
		log.Println(fmt.Sprintf("[debug] "+format, v...))
	}
}
