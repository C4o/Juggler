package golua

import (
	"regexp"

	"github.com/yuin/gopher-lua"
)

var reFns = map[string]lua.LGFunction{
	"find":  find,
	"match": match,
}

var cacheRec = make(map[string]*regexp.Regexp)

// 缓存re
func compile(repr string) (rec *regexp.Regexp, err error) {

	rec, err = regexp.Compile(repr)
	cacheRec[repr] = rec
	return rec, err
}

// 匹配结果返回 string, error
func find(L *lua.LState) int {

	var err error
	var rec *regexp.Regexp
	var ok bool
	str := L.CheckString(1)
	repr := L.CheckString(2)
	if rec, ok = cacheRec[repr]; ok {
		L.Push(lua.LString(rec.FindString(str)))
		L.Push(lua.LNil)
		return 2
	}
	rec, err = compile(repr)
	if err != nil {
		L.Push(lua.LString(""))
		pushErr(L, err)
	} else {
		L.Push(lua.LString(rec.FindString(str)))
		L.Push(lua.LNil)
	}
	return 2
}

// 匹配结果返回 bool, error
func match(L *lua.LState) int {

	var err error
	var rec *regexp.Regexp
	var ok bool
	str := L.CheckString(1)
	repr := L.CheckString(2)
	if rec, ok = cacheRec[repr]; ok {
		L.Push(lua.LBool(rec.MatchString(str)))
		L.Push(lua.LNil)
		return 2
	}
	rec, err = compile(repr)
	if err != nil {
		L.Push(lua.LBool(false))
		pushErr(L, err)
	} else {
		L.Push(lua.LBool(rec.MatchString(str)))
		L.Push(lua.LNil)
	}
	return 2
}

