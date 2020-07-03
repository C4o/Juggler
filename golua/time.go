package golua

import (
	"time"

	lua "github.com/yuin/gopher-lua"
)

var timeFns = map[string]lua.LGFunction{
	"unix":   tunix,
	"format": tformat,
	"server_date": server_date,
}

func tunix(L *lua.LState) int {

	L.Push(lua.LNumber(time.Now().Unix()))
	return 1
}

func tformat(L *lua.LState) int {

	L.Push(lua.LString(time.Now().Format("2006-01-02 15:04:05")))
	return 1
}

func server_date(L *lua.LState) int {

	L.Push(lua.LString(time.Now().Format("Mon, 02 Jan 2006 15:04:05 GMT")))
	return 1
}

