package golua

import (
	"time"

	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
)

var reqFns = map[string]lua.LGFunction{

}

func getHeader(L *lua.LState) int {

	L.Push(lua.LString(L.Context().(*gin.Context).GetHeader(L.CheckString(1))))
	return 1
}

func getReqVar(L *lua.LState) int {

	access, _ := L.Context().(*gin.Context)
	//logger.Printer(logger.ERROR, "ok is : %v", ok)
	r := access.Request
	w := access.Writer
	_ = L.CheckAny(1)
	switch L.CheckString(2) {
	case "host":
		L.Push(lua.LString(r.Host))
	case "status":
		L.Push(lua.LNumber(w.Status()))
	case "xff":
		L.Push(lua.LString(r.Header.Get("x-forwarded-for")))
	case "rule":
		L.Push(lua.LString(r.Header.Get("rule")))
	case "size":
		L.Push(lua.LNumber(w.Size()))
	case "method":
		L.Push(lua.LString(r.Method))
	case "uri":
		L.Push(lua.LString(r.URL.Path))
	case "app":
		L.Push(lua.LString(r.Header.Get("x-Rock-APP")))
	case "addr":
		L.Push(lua.LString(r.Header.Get("x-real-ip")))
	case "saddr":
		L.Push(lua.LString(r.RemoteAddr))
	case "query":
		L.Push(lua.LString(r.URL.RawQuery))
	case "ref":
		L.Push(lua.LString(r.Referer()))
	case "ua":
		L.Push(lua.LString(r.UserAgent()))
	case "ltime":
		L.Push(lua.LNumber(time.Now().Unix()))
	default:
		L.Push(lua.LNil)
	}
	return 1
}
