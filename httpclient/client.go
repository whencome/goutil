package httpclient

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/whencome/gotil"
	"github.com/whencome/xlog"
)

// HttpClient 定义一个包装了的HTTP Client，用于封装Http的相关操作
type HttpClient struct {
	Client      *http.Client    // 定义一个客户端
	Headers     gotil.M         // headers信息
	Cookies     []Cookie        // Cookie信息
	Params      gotil.M         // 请求参数列表，注意，这里的参数值是数组
	RawParam    []byte          // 原始请求数据
	Transport   *http.Transport // 传输协议
	Timeout     int             // 超时时间，单位：秒
	enableDebug bool            // 是否开启debug模式
}

// Cookie 定义一个基本的Cookie信息
type Cookie struct {
	Name     string
	Value    gotil.O
	HttpOnly bool
}

// NewHttpClient 创建一个http client对象
func NewHttpClient() *HttpClient {
	return &HttpClient{
		// Client:  &http.Client{},
		Headers:   gotil.M{},
		Cookies:   make([]Cookie, 0),
		Params:    gotil.M{},
		RawParam:  nil,
		Transport: nil,
		Timeout:   30, // 默认30秒超时,可以支持全局配置
	}
}

// 记录日志信息
func (client *HttpClient) Log(format string, data ...interface{}) {
	if !client.enableDebug {
		return
	}
	xlog.Debugf(format, data...)
}

// SetDebug 设置调试模式,true-表示开启调试，会打印请求信息, false-关闭调试模式
func (client *HttpClient) SetDebug(flag bool) {
	client.enableDebug = flag
}

// SetTransport 设置传输协议
func (client *HttpClient) SetTransport(t *http.Transport) {
	if t != nil {
		client.Transport = t
	}
}

// SetTimeout 设置超时时间
func (client *HttpClient) SetTimeout(timeout int) {
	if timeout <= 0 {
		timeout = 30
	}
	client.Timeout = timeout
}

// SetHeaders 设置头信息，设置之前会清空原有的头信息
func (client *HttpClient) SetHeaders(headers map[string]interface{}) {
	client.Headers = gotil.M{}
	client.AddHeaders(headers)
}

// AddHeaders 添加多个头信息（会覆盖已经存在的值）
func (client *HttpClient) AddHeaders(headers map[string]interface{}) {
	if len(headers) == 0 {
		return
	}
	for k, v := range headers {
		client.Headers[k] = v
	}
}

// AddHeader 添加头信息（会覆盖已经存在的值）
func (client *HttpClient) AddHeader(name string, val interface{}) {
	client.Headers[name] = val
}

// SetRawParam 设置原始请求参数
func (client *HttpClient) SetRawParam(param []byte) {
	client.RawParam = param
}

// SetParams 设置参数，设置之前会清空原有的参数
func (client *HttpClient) SetParams(params interface{}) {
	client.Params = gotil.M{}
	client.AddParams(params)
}

// AddParams 添加多个参数
func (client *HttpClient) AddParams(params interface{}) {
	switch params.(type) {
	case string:
		client.SetRawParam([]byte(params.(string)))
	default:
		kvParams := gotil.MVal(params)
		if kvParams == nil || len(kvParams) == 0 {
			return
		}
		for k, v := range kvParams {
			client.AddParam(k, gotil.SVal(v))
		}
	}
}

// AddParam 添加单个参数
func (client *HttpClient) AddParam(name string, val interface{}) {
	if _, ok := client.Params[name]; !ok {
		client.Params[name] = make([]*gotil.O, 0)
	}
	vals := gotil.SVal(val)
	if len(vals) == 0 {
		return
	}
	originVals := client.Params.SVal(name)
	originVals = append(originVals, vals...)
	client.Params[name] = originVals
}

// SetParam 以覆盖的方式添加参数，如果参数已经存在，其值会被覆盖
func (client *HttpClient) SetParam(name string, val interface{}) {
	if gotil.IsNil(val) {
		client.Params.Del(name)
	}
	client.Params[name] = make([]*gotil.O, 0)
	client.AddParam(name, val)
}

// AddCookie 添加单个参数
func (client *HttpClient) AddCookie(name string, val interface{}) {
	cookie := Cookie{
		Name:     name,
		Value:    gotil.Object(val),
		HttpOnly: true, // 此值暂时固定写死
	}
	client.Cookies = append(client.Cookies, cookie)
}

// buildRequest 构造请求信息
func (client *HttpClient) buildRequest(method, url string) (*http.Request, error) {
	queryString := client.buildQueryString()
	var body io.Reader
	if queryString != "" {
		if method == "GET" {
			if strings.Index(url, "?") >= 0 {
				url += "&"
			} else {
				url += "?"
			}
			url += queryString
		} else if method == "POST" {
			body = strings.NewReader(queryString)
		}
	}

	// 打印请求信息
	client.Log("%s %s\n", method, url)
	// 打印头信息
	for k, v := range client.Headers {
		client.Log("%s: %s\n", k, v)
	}
	client.Log("\n")
	// 打印请求内容
	client.Log("%s", queryString)
	client.Log("\n")

	// 构造请求
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	// 设置头信息
	if len(client.Headers) > 0 {
		for headKey, headVal := range client.Headers {
			request.Header.Set(headKey, gotil.String(headVal))
		}
	}
	// 添加cookie信息
	if len(client.Cookies) > 0 {
		for _, c := range client.Cookies {
			request.AddCookie(&http.Cookie{Name: c.Name, Value: c.Value.String(), HttpOnly: c.HttpOnly})
		}
	}
	// 设置http协议版本
	request.Proto = "HTTP/1.1"
	// 返回请求信息
	return request, nil
}

// buildQueryString 构造query string
func (client *HttpClient) buildQueryString() string {
	// 如果有raw param，则优先使用rawparam
	if client.RawParam != nil && len(client.RawParam) > 0 {
		return string(client.RawParam)
	}
	// 根据设置的参数进行处理
	if len(client.Params) == 0 {
		return ""
	}
	urlValues := client.Params.UrlValues()
	return urlValues.Encode()
}

// Get 执行http get请求
func (client *HttpClient) Get(url string) (*http.Response, error) {
	// 初始化http client
	client.Client = &http.Client{
		Timeout: time.Second * time.Duration(client.Timeout),
	}
	if client.Transport != nil {
		client.Client.Transport = client.Transport
	}
	// 构造请求
	request, err := client.buildRequest("GET", url)
	if err != nil {
		return nil, err
	}
	// 执行请求
	return client.Client.Do(request)
}

// Post 执行http post请求
func (client *HttpClient) Post(url string) (*http.Response, error) {
	// 添加默认头信息，防止post方式服务器端收不到数据
	if _, ok := client.Headers["Content-Type"]; !ok {
		client.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	}
	// 初始化http client
	client.Client = &http.Client{
		Timeout: time.Second * time.Duration(client.Timeout),
	}
	if client.Transport != nil {
		client.Client.Transport = client.Transport
	}
	// 构造请求
	request, err := client.buildRequest("POST", url)
	if err != nil {
		return nil, err
	}
	// 执行请求
	return client.Client.Do(request)
}
