package golua

import (
	"Juggler/logger"

	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
)

// 注册Request到lua虚拟机
func registerRock(L *lua.LState) {

	var rmt = L.NewTypeMetatable("req")
	var hmt = L.NewTypeMetatable("html")
	var reqTb, ginTb, htmlTb = &lua.LTable{}, &lua.LTable{}, &lua.LTable{}
	// 设置req的表和元表
	L.SetFuncs(rmt, map[string]lua.LGFunction{"__index": getReqVar})
	reqTb.Metatable = rmt
	ginTb.RawSet(lua.LString("var"), reqTb)
	// 设置logging方法和日志等级
	ginTb.RawSet(lua.LString("log"), L.NewFunction(logging))
	ginTb.RawSet(lua.LString("ERROR"), lua.LNumber(logger.ERROR))
	ginTb.RawSet(lua.LString("DEBUG"), lua.LNumber(logger.DEBUG))
	ginTb.RawSet(lua.LString("INFO"), lua.LNumber(logger.INFO))
	// 设置resp方法
	ginTb.RawSet(lua.LString("resp"), L.SetFuncs(L.NewTable(), respFns))
	// 设置req方法
	ginTb.RawSet(lua.LString("req"), L.SetFuncs(L.NewTable(), reqFns))
	htmlTb.Metatable = hmt
	ginTb.RawSet(lua.LString("html"), htmlTb)
	// 设置全局顶级变量table
	L.SetGlobal("rock", ginTb)
}

// 注册gin.context的userdata区域并填充日志数据
func registerGinContextUserData(L *lua.LState, c *gin.Context) {
	// 直接回传*gin.Context，类型就是context
	L.SetContext(c)
}

// 注册re相关方法到lua虚拟机
func luaRe(L *lua.LState) int {
	// 注册方法
	mod := L.SetFuncs(L.NewTable(), reFns)
	L.Push(mod)
	return 1
}

// 注册time相关方法到lua虚拟机
func luaTime(L *lua.LState) int {
	// 注册方法
	mod := L.SetFuncs(L.NewTable(), timeFns)
	mod.RawSet(lua.LString("zero"), lua.LNumber(1590829200))
	L.Push(mod)
	return 1
}

// 注册加密相关方法
func luaCrypto(L *lua.LState) int {
	// 注册方法
	mod := L.SetFuncs(L.NewTable(), cryptoFns)
	L.Push(mod)
	return 1
}

// 注册随机数方法
func luaRandom(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), randFns)
	L.Push(mod)
	return 1
}
