local var = rock.var
local resp = rock.resp
local crypto = require("crypto")
local time = require("time")
local re = require("re")
local log = rock.log
local ERR = rock.ERROR

if var.rule == "sqli" then
    resp.string(200, "Congratulation！Password hash is %s.", crypto.randomMD5(32))
    log(ERR, "found sqli attack in %d", time.format())
    return
end

if var.rule == "xss" and re.match(var.uri, "^/admin/") then
    resp.set_header("Content-Type", "text/html; charset=utf-8")
    resp.set_header("Date", time.server_date())
    resp.status(403)
    return
end

if var.rule == "lfi_shadow" then
    resp.html(200, "etc_shadow")
    return
end

if var.rule == "rce" then
    resp.set_header("Content-Type", "text/html; charset=utf-8")
    resp.set_cookie("sessionid", "admin_session", 6000, "/", var.host, true, true)
    resp.clone(200, "https://duxiaofa.baidu.com/detail?searchType=statute&from=aladdin_28231&originquery=%E7%BD%91%E7%BB%9C%E5%AE%89%E5%85%A8%E6%B3%95&count=79&cid=f66f830e45c0490d589f1de2fe05e942_law")
    return
end

