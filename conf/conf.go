package conf

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	Cfg        Config
	Files      map[string][]string // 文件夹内的文件，map[文件夹名]文件名数组
	Headers    map[string]map[string]string
	ConfigFile string
)

type Default struct {
	B64    bool
	Html   string // 匹配不到策略后默认拦截返回的html目录
	Status int    // 匹配不到策略后默认拦截的状态码,优先级低于Html
	Debug  int    // 日志打印等级
}

type Incept struct {
	Html    string
	Headers string
	Status  int
}

type Kafka struct {
	Addr   string
	Thread int
	Num    int
	Topic  string
	Local  string
}

type Config struct {
	// 默认拦截配置
	Default Default
	// 拦截配置
	Incept map[string]map[string]map[string]Incept
	// 请求体编码
	Encode map[string]bool
	// 爬虫配置
	Crawler map[string]map[string]map[string]string
	Kafka   Kafka
}

func (cfg *Config) Load() {

	if tomlFile, err := ioutil.ReadFile(ConfigFile); err == nil {
		*cfg = Config{}
		Files = make(map[string][]string)
		Headers = make(map[string]map[string]string)
		_, err := toml.Decode(string(tomlFile), &cfg)
		if err != nil {
			log.Printf("error in decode toml : %v", err)
		}
		Files = make(map[string][]string)
		Headers = make(map[string]map[string]string)
		log.Println("configuration update.")
	} else {
		log.Printf("error in open file : %v", err)
	}
}

func (cfg *Config) Monitor() {

	var last int64
	var f os.FileInfo
	var err error

	if f, err = os.Stat(ConfigFile); err == nil {
		last = f.ModTime().Unix()
		s1 := time.NewTicker(1 * time.Second)
		defer func() {
			s1.Stop()
		}()
		for {
			select {
			case <-s1.C:
				if f, err = os.Stat(ConfigFile); err == nil {
					if last != f.ModTime().Unix() {
						cfg.Load()
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

//func (cfg *Config) Parse(deny map[string]Incept) (int, string) {

//}

func (cfg *Config) ReConfig(ic map[string]Incept, uri string) Incept {

	for k, v := range ic {
		if result, err := regexp.MatchString(k, uri); err == nil && result {
			return v
		}
	}
	return Incept{}
}

func (cfg *Config) Return(host, uri, rule string) (int, string) {

	var status int
	var html, path string
	var err error

	if path = cfg.ReConfig(Cfg.Incept[host][rule], uri).Html; path == "" {
		// 没有静态文件，状态码存在。
		if status = cfg.ReConfig(Cfg.Incept[host][rule], uri).Status; status != 0 {
			return status, ""
		}
	} else {
		if html, err = cfg.GetHtml(path); err == nil {
			html = strings.Replace(html, "IS.MD5", genRandomMD5(time.Now().String()), -1)
			return 200, html
		}
	}

	// 静态文件和状态码都不存在
	if path = Cfg.Default.Html; path != "" {
		if html, err = Cfg.GetHtml(path); err == nil {
			html = strings.Replace(html, "IS.MD5", genRandomMD5(time.Now().String()), -1)
			return 200, html
		}
	}
	if status = Cfg.Default.Status; status != 0 {
		return status, ""
	} else {
		log.Printf("%s", Cfg)
		return 200, ""
	}
}

func (cfg *Config) GetHtml(path string) (string, error) {

	// 先读缓存
	if len(Files[path]) == 0 {
		// 把目标文件夹文件名存入缓存
		if rd, err := ioutil.ReadDir(path); err == nil {
			for _, fi := range rd {
				if !fi.IsDir() {
					Files[path] = append(Files[path], fi.Name())
				}
			}
		} else {
			return "", err
		}
	}

	// 读一个随机文件
	content, err := ioutil.ReadFile(path + Files[path][rand.Intn(len(Files[path]))])
	if err == nil {
		return string(content), nil
	}
	return "", err
}

func (cfg *Config) SetHeaders(host, uri, rule string) (map[string]string, error) {

	var inf map[string]string
	hs := cfg.ReConfig(Cfg.Incept[host][rule], uri).Headers
	if len(Headers[hs]) != 0 {
		return Headers[hs], nil
	} else {
		err := json.Unmarshal([]byte(hs), &inf)
		if err == nil {
			Headers[hs] = inf
		} else {
			return inf, err
		}
	}
	return Headers[hs], nil
}

func genRandomMD5(data string) string {

	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
