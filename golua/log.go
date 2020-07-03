package golua

import (
	"Juggler/logger"
	"bytes"

	lua "github.com/yuin/gopher-lua"
)

func logging(L *lua.LState) int {

	buf := new(bytes.Buffer)
	n := L.GetTop()

	for i := 2; i < n+1; i++ {
		buf.WriteString(L.CheckString(i))
		buf.WriteString(" ")}

	logger.Printer(L.CheckInt(1), buf.String())
	return 0
}
