package golua

import (
	"Juggler/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"io/ioutil"

	lua "github.com/yuin/gopher-lua"
)

func (pl *LStatePool) LoadHtml(path string) error {

	if path == "" {
		path = "html/"
	} else {
		pl.HPath = path
	}
	// 初始加载自定义相应内容
	fileList, err := ioutil.ReadDir(pl.HPath)
	if err != nil {
		logger.Printer(logger.ERROR, "load htmls in %s error : %v", pl.HPath, err)
		return err
	}
	for _, fileInfo := range fileList {
		name := fileInfo.Name()
		LuaPool.Htmls[name[:len(name)-5]], err = ioutil.ReadFile(pl.HPath+name)
		if err == nil {
			logger.Printer(logger.INFO, "update html file %s successfully.", name)
		} else {
			delete(LuaPool.Htmls, name[:len(name)-5])
			logger.Printer(logger.ERROR, "read html file %s error : %v", name, err)
		}
	}
	go LuaPool.MonitorHtml()
	return nil
}

func (pl *LStatePool) MonitorHtml()  {

	var watcher *fsnotify.Watcher
	var event fsnotify.Event
	var err error
	// 检测插件文件是否变化
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		logger.Printer(logger.ERROR, "new inotify watcher error: %v", err)
	}
	defer watcher.Close()
	watcher.Add(pl.HPath)
	for {
		select {
		case event =<- watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				if event.Name[len(event.Name)-5:] == ".html" {
					LuaPool.Htmls[event.Name[len(pl.HPath):len(event.Name)-5]], err = ioutil.ReadFile(event.Name)
					if err == nil {
						logger.Printer(logger.INFO, "update html file %s successfully.", event.Name)
					} else {
						delete(LuaPool.Htmls, event.Name[len(pl.HPath):len(event.Name)-5])
						logger.Printer(logger.ERROR, "read html file %s error : %v", event.Name, err)
					}
				} else {
					logger.Printer(logger.ERROR, "%s is not end with .html!", event.Name)
				}
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				delete(LuaPool.Htmls, event.Name[len(pl.HPath):len(event.Name)-5])
				logger.Printer(logger.INFO, "delete html file %s", event.Name)
			}
		}
	}
}

func getHtml(L *lua.LState) int {

	var html []byte
	var ok bool
	key := L.CheckString(2)
	if html, ok = LuaPool.Htmls[key]; !ok {
		logger.Printer(logger.ERROR, "cannot load file %s.", key)
		html = LuaPool.Htmls["default_404"]
	}
	n := L.GetTop()
	buf := make([]interface{}, n-2)
	for i := 3; i < n+1; i++ {
		buf[i-3] = L.CheckAny(i)
	}
	L.Context().(*gin.Context).String(L.CheckInt(1), string(html), buf...)
	return 0
}