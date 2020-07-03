package golua

import (
	"Juggler/config"
	"Juggler/logger"
	"bufio"
	"errors"
	"io/ioutil"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/ast"
	"github.com/yuin/gopher-lua/parse"
)

var (
	LuaPool *LStatePool
	workerVM = 1
	functionVM = 0
)

// lua虚拟机池
type LStatePool struct {
	HPath string
	Htmls map[string][]byte
	FnVM  *lua.LState
	Luas  string
	Fns   map[string]*lua.LFunction
	VMs chan *lua.LState
}

// 拿走channel中的虚拟机
func (pl *LStatePool) get() *lua.LState {
	n := len(pl.VMs)
	if n == 0 {
		return pl.new(workerVM)
	}
	x := <- pl.VMs
	return x
}

// 把用完的虚拟机放回channel
func (pl *LStatePool) put(L *lua.LState) {
	pl.VMs <- L
}

// 新建虚拟机，注册若干预设定变量和模块
func (pl *LStatePool) new(vmType int) *lua.LState {

	switch vmType {
	case functionVM:
		logger.Printer(logger.INFO, "fnVM new state")
	case workerVM:
		logger.Printer(logger.INFO, "workerVM new state")
	}
	L := lua.NewState()
	registerRock(L)
	L.PreloadModule("re", luaRe)
	L.PreloadModule("time", luaTime)
	L.PreloadModule("crypto", luaCrypto)
	L.PreloadModule("random", luaRandom)
	return L
}

// 结束后关闭所有channel中的虚拟机
func (pl *LStatePool) Shutdown() {
	for L := range pl.VMs {
		L.Close()
	}
}

// 初始化若干个虚拟机放入channel待用
func (pl *LStatePool) Init(s string) error {

	if s == "" {
		LuaPool.Luas = "scripts/"
	} else {
		if s[len(s)-1:] != "/" {
			logger.Printer(logger.ERROR, "%s plugins path must be end with '/' !", s)
			return errors.New("plugins path error.")
		}
		LuaPool.Luas = s
	}
	logger.Printer(logger.INFO, "vmnum is %d, init of struct...", config.Cfg.Other.VMNum)
	pl.FnVM = pl.new(functionVM)
	go pl.loadPlugins()
	for i := 0; i < config.Cfg.Other.VMNum; i++ {
		pl.VMs <- pl.new(workerVM)
	}
	return nil
}

// 获取方法供使用
func (pl *LStatePool) getFn(host string) *lua.LFunction {

	if fn, ok := pl.Fns[host]; ok {
		return fn
	}
	return pl.Fns["default"]
}

// 加载和规则
func (pl *LStatePool) loadPlugins() {

	var watcher *fsnotify.Watcher
	var event fsnotify.Event

	// 初始加载插件
	fileList, err := ioutil.ReadDir(pl.Luas)
	if err != nil {
		logger.Printer(logger.ERROR, "plugin path %s read error : %v", pl.Luas, err)
		os.Exit(1)
	}
	for _, fileInfo := range fileList {
		name := fileInfo.Name()
		if name[len(name)-4:] == ".lua" {
			pl.compileLua(pl.Luas+name,name[:len(name)-4])
		} else {
			logger.Printer(logger.ERROR, "%s is not end with .lua!", name)
		}
	}
	// 检测插件文件是否变化
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		logger.Printer(logger.ERROR, "new inotify watcher error: %v", err)
	}
	defer watcher.Close()
	watcher.Add(pl.Luas)
	for {
		select {
		case event =<- watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				if event.Name[len(event.Name)-4:] == ".lua" {
					pl.compileLua(event.Name, event.Name[len(pl.Luas):len(event.Name)-4])
				} else {
					logger.Printer(logger.ERROR, "%s is not end with .lua!", event.Name)
				}
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				delete(pl.Fns, event.Name[len(pl.Luas):len(event.Name)-4])
				logger.Printer(logger.INFO, "delete plugins %s", event.Name)
			}
		}
	}
}

// 编译lua脚本并缓存方法
func (pl *LStatePool) compileLua(filepath, host string) {

	var err error
	var file *os.File
	var chunk []ast.Stmt
	var proto *lua.FunctionProto

	file, err = os.OpenFile(filepath, os.O_RDONLY, 0444)
	if err != nil {
		logger.Printer(logger.ERROR, "script %s not found! error is %v.", filepath, err)
		return
	}
	defer file.Close()
	chunk, err = parse.Parse(bufio.NewReader(file), filepath)
	if err != nil {
		logger.Printer(logger.ERROR, "parse script %s failed for %v.", filepath, err)
		return
	}
	proto, err = lua.Compile(chunk, filepath)
	if err != nil {
		logger.Printer(logger.ERROR, "compile script %s failed for %v.", filepath, err)
		return
	}
	pl.Fns[host] = pl.FnVM.NewFunctionFromProto(proto)
	logger.Printer(logger.INFO, "update plugins %s successfull", host)
	return
}

// 每个请求的worker
func LuaWorker(c *gin.Context) {

	var err error
	L := LuaPool.get()
	defer LuaPool.put(L)
	// 每次将gin.context存入userdata待使用
   	registerGinContextUserData(L, c)
   	L.Push(LuaPool.getFn(c.Request.Host))
	err = L.PCall(0, lua.MultRet, nil)
   	if err != nil {
		logger.Printer(logger.ERROR, "%v", err)
	}
}