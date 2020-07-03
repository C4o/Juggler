package golua

import (
	jsoniter "github.com/json-iterator/go"
	lua "github.com/yuin/gopher-lua"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func pushErr(L *lua.LState, err error) {

	if err == nil {
		L.Push(lua.LNil)
	} else {
		L.Push(lua.LString(err.Error()))
	}
}

func typeCheck(it interface{}) lua.LValue {

	var ok bool
	if it == nil {
		return lua.LNil
	} else if _, ok = it.(string); ok {
		return lua.LString(it.(string))
	} else if _, ok = it.(int); ok {
		return lua.LNumber(it.(int))
	} else {
		return lua.LNumber(1)
	}
}
