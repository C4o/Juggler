package web

import (
	"Juggler/golua"

	"github.com/gin-gonic/gin"
)

// gin路由处理函数
func handler(c *gin.Context) {

	golua.LuaWorker(c)

	//c.String(200, string(cjpl))
}

