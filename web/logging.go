package web

import (
	"Juggler/config"
	"Juggler/logger"
	"encoding/base64"
	"net"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oschwald/geoip2-golang"
)

type Buff struct {
	Ch   chan Requests
	Size int
	File *os.File
}

func Logging() gin.HandlerFunc {

	//logger.Printer(logger.ERROR, "logging")
	return func(c *gin.Context) {

		var err error
		var body []byte
		var record *geoip2.City
		geodata := GeoData{}
		r := c.Request
		w := c.Writer

		req := Requests{
			TimeStamp: time.Now().Format("02/Jan/2006:15:04:05 +0800"),
			Saddr:     r.RemoteAddr,
			Method:    r.Method,
			Host:      r.Host,
			UA:        r.UserAgent(),
			URI:       r.URL.Path,
			Query:     r.URL.RawQuery,
			Rule:      r.Header.Get("rule"),
			XFF:       r.Header.Get("x-forwarded-for"),
			REF:       r.Referer(),
			Addr:      r.Header.Get("x-real-ip"),
			APP:       r.Header.Get("x-Rock-APP"),
			Headers:   r.Header,
			Status:    w.Status(),
			Size:      w.Size(),
		}

		body, err = c.GetRawData()
		//logger.Printer(logger.DEBUG, "addr is %v", req.Addr)
		if err != nil {
			logger.Printer(logger.ERROR, "get raw data error : %v", err)
		} else {
			// 获取地理位置信息
			if len(req.Addr) > 0 {
				loc := Loc.Find(req.Addr)
				req.LRegion = loc.Region
				req.LCountry = loc.Country
				req.LProvince = loc.Province
				req.LCity = loc.City
			}
			req.Body = base64.StdEncoding.EncodeToString(body)
			req.Local = config.Cfg.Other.Local

			record, err = GeoFile.City(net.ParseIP(req.Addr))
			if err != nil {
				logger.Printer(logger.ERROR, "get geoip data error : %v", err)
			} else {
				geodata.Lat = record.Location.Latitude
				geodata.Lon = record.Location.Longitude
				req.Location = geodata
				//fmt.Printf("geodata : %v", geodata)
			}
			// 如果开关开启，就往kafka的channel里传
			if config.Cfg.Kafka.On {
				KAccess.Chan <- req
			}
		}
	}
}
