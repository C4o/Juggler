package web

import (
	"IUS/conf"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oschwald/geoip2-golang"
)

var (
	Kafka = KafkaAccess{Chan: make(chan requests, 40960*2)}
)

type requests struct {
	TimeStamp string // 访问时间
	Addr      string // 来源地址，一般是WAF地址
	Host      string // host头，区分业务
	UA        string
	URI       string
	Query     string
	Rule      string              // 匹配上的规则，便于分类拦截
	XFF       string              // X-Forwarded-For
	XRI       string              // X-Real-IP
	Method    string              // 请求方式
	APP       string              // 所属应用
	Headers   map[string][]string // 完整头数据
	Status    int                 // 响应状态码
	Size      int                 // 响应包长度
	Body      string              // post body全包
	LRegion   string              // 组织
	LCountry  string              // 国家
	LProvince string              // 省份
	LCity     string              // 城市
	LCounty   string              // 镇
	Local     string              // 拦截中心地址
	Location  GeoData             // 经纬度
}

type Buff struct {
	Ch   chan requests
	Size int
	File *os.File
}

type GeoData struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func Logger() gin.HandlerFunc {

	return func(c *gin.Context) {

		var err error
		var body []byte
		var record *geoip2.City
		geodata := GeoData{}
		r := c.Request
		c.Next()
		w := c.Writer

		req := requests{
			TimeStamp: time.Now().Format("02/Jan/2006:15:04:05 +0800"),
			Addr:      r.RemoteAddr,
			Method:    r.Method,
			Host:      r.Host,
			UA:        r.UserAgent(),
			URI:       r.URL.Path,
			Query:     r.URL.RawQuery,
			Rule:      r.Header.Get("rule"),
			XFF:       r.Header.Get("x-forwarded-for"),
			XRI:       r.Header.Get("x-real-ip"),
			APP:       r.Header.Get("x-waf-APP"),
			Headers:   r.Header,
			Status:    w.Status(),
			Size:      w.Size(),
		}

		body, err = c.GetRawData()
		if err != nil {
			log.Printf("get raw data error : %v", err)
		} else {
			// 获取地理位置信息
			if len(req.XRI) > 0 {
				loc := Loc.Find(req.XRI)
				req.LRegion = loc.Region
				req.LCountry = loc.Country
				req.LProvince = loc.Province
				req.LCity = loc.City
				req.LCounty = loc.County
			}

			req.Body = B64(body, req.Host)
			req.Local = conf.Cfg.Kafka.Local

			record, err = GeoFile.City(net.ParseIP(req.XRI))
			if err != nil {
				log.Printf("get geoip data error : %v", err)
			} else {
				geodata.Lat = record.Location.Latitude
				geodata.Lon = record.Location.Longitude
				req.Location = geodata
				//fmt.Printf("geodata : %v", geodata)
			}

			Kafka.Chan <- req
		}

	}
}

func B64(body []byte, host string) string {
	if conf.Cfg.Encode[host] {
		return base64.StdEncoding.EncodeToString(body)
	}
	return fmt.Sprintf("%s", body)
}
