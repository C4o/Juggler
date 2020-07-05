# Juggler - 一个也许能骗到黑客的系统

## 应用场景
现在很多WAF拦截了恶意请求之后，直接返回一些特殊告警页面（之前有看到t00ls上有看图识WAF）或一些状态码（403或者500啥的）。

但是实际上返不返回特殊响应都不会有啥实际作用，反而会给攻击者显而易见的提示。

但是如果返回的内容跟业务返回一致的话，就能让攻击者很难察觉到已经被策略拦截了。

场景一：攻击者正在暴力破解某登陆口
```
发现登陆成功是
{"successcode":0,"result":{"ReturnCode":0}}
登陆失败是
{"errorcode":1,"error":"用户名密码不匹配","result":{"ReturnCode":0}}
```
那现在我们可以这么做
```
1. 触发规则后持续返回错误状态码，让黑客觉得自己的字典不大行。
2. 返回一个特定的cookie，当waf匹配到该cookie后，将请求导流到某web蜜罐跟黑客深入交流。
```

场景二：攻击者正在尝试找xss

我们可以这么做
```
例如：
1. 不管攻击者怎么来，检测后都返回去去除了攻击者payload的请求的响应。
2. 攻击者payload是alert(xxxx)，那不管系统有无漏，我们返回一个弹框xxx。
  （当然前提是我们能识别payload的语法是否正确，也不能把攻击者当傻子骗。）
```

肯定有人会觉得，我们WAF强的不行，直接拦截就行，不整这些花里胡哨的，那这可以的。

但是相对于直接的拦截给攻击者告警，混淆视听，消费攻击者的精力，让攻击者怀疑自己，这样是不是更加狡猾？这也正是项目取名的由来，juggler，耍把戏的人。

当然，上面需求实现的前提，是前方有一个强有力的WAF，只有在攻击请求被检出后，攻击请求才能到达我们的拦截欺骗中心，否则一切都是扯犊子。

```
项目思路来自我的领导们，并且简单的应用已经在线上有了很长一段时间的应用，我只是思路的实现者。
项目已在线上运行一年多，每日处理攻击请求过亿。

juggler本质上是一个lua插件化的web服务器，类似openresty（大言不惭哈哈）；
基于gin进行的开发，其实就是将*gin.Context以lua的userdata放入lua虚拟机，所以可以通过lua脚本进行请求处理。
```

### 性能

跟gin进行对比，性能损失大概10%。

虽然每个请求的真实处理还是在golang中完成，但是每个请求的一些临时变量都会在lua虚拟机走一遍。

gin逻辑
```go
func handler(c *gin.Context) {
    c.String(200, "host of this request is %s", c.Request.Host)
}
```
juggler逻辑
```lua
local var = rock.var
local resp = rock.resp

resp.string(200, "host of this request is %s", var.host)
```

### 使用方式

项目流程图

示例插件

```lua
-- juggler.test.com.lua
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
```

## 特点

### 插件编写灵活

### 插件和响应文件被动式更新

### 丰富三方插件库可自行定义

## 已实现需求

### 功能

### 每个请求可操作的变量

### 内置模块和对应需求

## 联动WAF使用流程

## 本项目在现实中的应用

### WAF体系

本项目为拦截图中的拦截欺骗中心，接收并处理所有恶意请求。

![image](https://p3.ssl.qhimg.com/t015b7079b7b1839010.png)

### 日志分析风控系统

项目地址：[https://github.com/C4o/FBI-Analyzer](https://github.com/C4o/FBI-Analyzer)

### 实时日志传输模块

项目地址：[https://github.com/C4o/LogFarmer](https://github.com/C4o/LogFarmer)

