package golua

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
)

var cryptoFns = map[string]lua.LGFunction{
	"md5sum":    md5sum,
	"randomMD5": randomMD5,
	"b64encode": b64encode,
	"b64decode": b64decode,
}

func md5sum(L *lua.LState) int {

	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%+v", L.CheckAny(1))))
	L.Push(lua.LString(hex.EncodeToString(h.Sum(nil))))
	return 1
}

func randomMD5(L *lua.LState) int {

	h := md5.New()
	h.Write([]byte(fmt.Sprintf("x%d", time.Now().Nanosecond())))
	if L.CheckInt(1) == 16 {
		L.Push(lua.LString(hex.EncodeToString(h.Sum(nil))[8:24]))
	} else {
		L.Push(lua.LString(hex.EncodeToString(h.Sum(nil))))
	}

	return 1
}

func b64encode(L *lua.LState) int {

	return 0
}

func b64decode(L *lua.LState) int {

	return 0
}