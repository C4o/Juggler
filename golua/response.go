package golua

import (
	"Juggler/logger"

	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
)

var respFns = map[string]lua.LGFunction{
	"string":     respString,
	"status":     respStatus,
	"html":       getHtml,
	"clone":      clone,
	"set_header": respSetHeader,
	"set_cookie": respSetCookie,
}

func respString(L *lua.LState) int {

	status := L.CheckInt(1)
	format := L.CheckString(2)
	n := L.GetTop()
	buf := make([]interface{}, n-2)
	for i := 3; i < n+1; i++ {
		buf[i-3] = L.CheckAny(i)
	}
	L.Context().(*gin.Context).String(status, format, buf...)
	return 0
}

func respStatus(L *lua.LState) int {

	logger.Printer(logger.ERROR, "%d", L.GetTop())
	L.Context().(*gin.Context).Status(L.CheckInt(1))
	return 0
}

func respSetHeader(L *lua.LState) int {

	L.Context().(*gin.Context).Header(L.CheckString(1), L.CheckString(2))
	return 0
}

func respSetCookie(L *lua.LState) int {

	L.Context().(*gin.Context).SetCookie(
		L.CheckString(1), L.CheckString(2), L.CheckInt(3), L.CheckString(4),
		L.CheckString(5), L.CheckBool(6), L.CheckBool(7))
	return 0
}
