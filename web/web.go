package web

import (
	"Juggler/iploc"
	"Juggler/logger"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/gzip"
	"github.com/oschwald/geoip2-golang"
)

var (
	GeoFile *geoip2.Reader
	Loc     *iploc.Locator
	NoGzip = 0
	DefaultGzip = 1
	BestGzip = 2
)

func WebServer(addr string, gzipType int) {

	var err error
	// 初始化纯真数据库
	Loc, err = iploc.Open("qqwry-0.dat")
	if err != nil {
		logger.Printer(logger.ERROR, "open qqwry.dat error : %v", err)
		return
	}
	// 初始化geoip数据，方便kibana做攻击来源热力图
	GeoFile, err = geoip2.Open("GeoIP.mmdb")
	if err != nil {
		logger.Printer(logger.ERROR, "open GeoIP.mmdb error : %v", err)
		return
	}
	defer GeoFile.Close()

	// 启动gin
	r := gin.New()
	gin.SetMode(gin.ReleaseMode)
	switch gzipType {
	case NoGzip:
		r.Use(gzip.Gzip(gzip.NoCompression))
	case DefaultGzip:
		r.Use(gzip.Gzip(gzip.DefaultCompression))
	case BestGzip:
		r.Use(gzip.Gzip(gzip.BestCompression))
	}
	r.Use(gin.Recovery(), Logging())
	r.NoRoute(handler)
	//r.GET("/", Handler)
	r.Run(addr)
}
