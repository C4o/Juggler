package golua

import (
	lua "github.com/yuin/gopher-lua"
	"math/rand"
	"time"
)

var randFns = map[string]lua.LGFunction{
	"rint" : rint,
}

func rint(L *lua.LState) int {

	rand.Seed(time.Now().UnixNano())
	L.Push(lua.LNumber(rand.Intn(L.CheckInt(1))+1))
	return 1
}