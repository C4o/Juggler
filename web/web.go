package web

import (
	"IUS/iploc"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/oschwald/geoip2-golang"
)

var (
	GeoFile *geoip2.Reader
	Loc     *iploc.Locator
	F       *os.File
)

func WebServer(addr string) {

	var err error
	Loc, err = iploc.Open("qqwry-0.dat")
	if err != nil {
		log.Printf("open qqwry.dat error : %v", err)
		return
	}
	GeoFile, err = geoip2.Open("GeoIP.mmdb")
	if err != nil {
		log.Printf("open GeoIP.mmdb error : %v", err)
		return
	}
	defer GeoFile.Close()

	r := gin.New()
	gin.SetMode(gin.ReleaseMode)
	r.Use(gin.Recovery(), Logger())
	r.NoRoute(handler)
	r.Static("/static", "./static")
	r.Run(addr)
}
