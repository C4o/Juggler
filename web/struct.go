package web

type Requests struct {
	TimeStamp string              `json:"timestamp"` // 访问时间
	Saddr     string              `json:"saddr"`     // 来源地址，一般是WAF地址
	Host      string              `json:"host"`      // host头，区分业务
	UA        string              `json:"ua"`        // UA头
	URI       string              `json:"uri"`       // uri
	Query     string              `json:"query"`     // uri里的查询参数
	Rule      string              `json:"rule"`      // 匹配上的规则，便于分类拦截
	XFF       string              `json:"xff"`       // X-Forwarded-For
	REF       string              `json:"ref"`       // referer
	Addr      string              `json:"addr"`      // X-Real-IP
	Method    string              `json:"method"`    // 请求方式
	APP       string              `json:"app"`       // 所属应用
	Headers   map[string][]string `json:"headers"`   // 完整头数据
	Status    int                 `json:"status"`    // 响应状态码
	Size      int                 `json:"size"`      // 响应包长度
	Body      string              `json:"body"`      // post body全包
	LRegion   string              `json:"region"`    // 组织
	LCountry  string              `json:"country"`   // 国家
	LProvince string              `json:"province"`  // 省份
	LCity     string              `json:"city"`      // 城市
	Local     string              `json:"localip"`   // 拦截中心地址
	Location  GeoData             `json:"Location"`  // 经纬度
}

type GeoData struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}
