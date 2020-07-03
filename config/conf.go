package config

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

var Cfg Config

type Kafka struct {
	Addr   string
	Thread int
	Num    int
	Topic  string
	// 开关，决定是否传到kafka里
	On     bool
}

type Other struct {
	// 拦截节点地址
	Local string
	// 日志等级
	Debug int
	// lua虚拟机个数
	VMNum int
}

type Config struct {
	// kafka配置
	Kafka Kafka
	Other Other
}

func (cfg *Config) Load(ConfPath string) error {

	if tomlFile, err := ioutil.ReadFile(ConfPath); err == nil {
		*cfg = Config{}
		_, err := toml.Decode(string(tomlFile), &cfg)
		if err != nil {
			log.Printf("error in decode toml : %v", err)
		}
		log.Println("[info] configuration update.")
		return nil
	} else {
		log.Printf("[error] error in open file : %v", err)
		return err
	}
}

func (cfg *Config) Monitor(ConfPath string) {

	var last int64
	var f os.FileInfo
	var err error

	if f, err = os.Stat(ConfPath); err == nil {
		last = f.ModTime().Unix()
		s1 := time.NewTicker(1 * time.Second)
		defer func() {
			s1.Stop()
		}()
		for {
			select {
			case <-s1.C:
				if f, err = os.Stat(ConfPath); err == nil {
					if last != f.ModTime().Unix() {
						cfg.Load(ConfPath)
						last = f.ModTime().Unix()
					}
				} else {
					log.Printf("stat file error : %v", err)
				}
			}
		}
	} else {
		log.Printf("stat file error : %v", err)
	}
}