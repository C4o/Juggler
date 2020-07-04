-- 文件名juggler.test.com.lua 当攻击请求的业务域名是juggler.test.com时匹配该插件
local var = rock.var
local resp = rock.resp
local crypto = require("crypto")
local time = require("time")
local re = require("re")
local log = rock.log
local ERR = rock.ERROR

-- 通过var内的参数，匹配每一个攻击请求中的http参数
if var.rule == "sqli" then
    -- 满足条件后直接返回格式化字符串，使用内置方法每次回显不同的32位随机md5值
    resp.string(200, "Congratulation！Password hash is %s.", crypto.randomMD5(32))
    -- 在日志文件中打印日志
    log(ERR, "found sqli attack in %d", time.format())
    return
end

-- 使用正则匹配某个路径，与规则匹配并用
if var.rule == "xss" and re.match(var.uri, "^/admin/") then
    -- 设置响应体类型
    resp.set_header("Content-Type", "text/html; charset=utf-8")
    -- 添加响应头Date，内容是正常服务器产生的内容
    resp.set_header("Date", time.server_date())
    -- 只响应状态码，不响应内容
    resp.status(403)
    return
end

if var.rule == "lfi_shadow" then
    -- 使用预存文件etc_shadow.html进行内容回显，状态码200
    resp.html(200, "etc_shadow")
    return
end

if var.rule == "rce" then
    resp.set_header("Content-Type", "text/html; charset=utf-8")
    -- 在响应中set_cookie
    resp.set_cookie("sessionid", "admin_session", 6000, "/", var.host, true, true)
    -- 克隆固定页面回显，缓存内容，不会每次都克隆
    resp.clone(200, "https://duxiaofa.baidu.com/detail?searchType=statute&from=aladdin_28231&originquery=%E7%BD%91%E7%BB%9C%E5%AE%89%E5%85%A8%E6%B3%95&count=79&cid=f66f830e45c0490d589f1de2fe05e942_law")
    return
end

-- 不匹配任何规则时，返回默认404内容
resp.set_header("Content-Type", "text/html; charset=utf-8")
resp.html(404, "default_404")
return