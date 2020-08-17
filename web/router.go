package web

import (
	"IUS/conf"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

func handler(c *gin.Context) {

	rule := c.Request.Header.Get("Rule")
	host := c.Request.Host
	uri := c.Request.RequestURI

	status, html := conf.Cfg.Return(host, uri, rule)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if headers, err := conf.Cfg.SetHeaders(host, uri, rule); err == nil {
		for k, v := range headers {
			if v == "NowTime" {
				v = fmt.Sprintf("%s, %s GMT", time.Now().Weekday().String(), time.Now().Format("02 Jan 2006 15:04:05"))
			}
			c.Header(k, v)
		}
	}
	c.String(status, "%s", html)
}
