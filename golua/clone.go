package golua

import (
	"Juggler/logger"
	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
	"io/ioutil"
	"net/http"
)

var (
	cloneSites = make(map[string][]byte)
	client = &http.Client{}
)

func clone(L *lua.LState) int {

	var content []byte
	var ok bool
	var url string
	status := L.CheckInt(1)
	url = L.CheckString(2)
	content, ok = cloneSites[url]
	if !ok {
		go goCloneSite(url)
		content = LuaPool.Htmls["default"]
	}
	L.Context().(*gin.Context).String(status, string(content))
	return 0
}

func goCloneSite(url string)  {

	var err error
	var resp *http.Response
	var body []byte
	req, err := http.NewRequest("GET", url, nil) //建立一个请求
	if err != nil {
		logger.Printer(logger.ERROR, "cannot new http request to %s , error is %v", url, err)
		return
	}
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Language", "ja,zh-CN;q=0.8,zh;q=0.6")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")
	resp, err = client.Do(req) //提交
	defer resp.Body.Close()
	if err != nil {
		logger.Printer(logger.ERROR, "cannot send http request to %s , error is %v", url, err)
		return
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Printer(logger.ERROR, "cannot read http response of %s , error is %v", url, err)
		return
	}
	cloneSites[url] = body
}
