-- 当攻击请求无法匹配到特定插件时，匹配默认插件
local resp = rock.resp

resp.set_header("Content-Type", "text/html; charset=utf-8")
resp.html(404, "default_404")
return