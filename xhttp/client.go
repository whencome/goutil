package xhttp

import (
    "errors"
    "io"
    "net/http"
    "net/url"
    "reflect"
    "strings"
    "time"
)

var (
    ErrInvalidParamType = errors.New("can request with both key-value params and stream params")
)

const (
    paramTypeNone   = 0
    paramTypeKV     = 1
    paramTypeStream = 2
)

// clientOption 定义设置客户端参数的option对象
type clientOption func(c *Client)

// Client 定义一个包装了的HTTP Client，用于封装Http的相关操作
type Client struct {
    Client    *http.Client      // 定义一个客户端
    Headers   map[string]string // headers信息
    Cookies   []Cookie          // Cookie信息
    Params    url.Values        // 请求参数列表
    RawParam  []byte            // 原始请求数据
    Body      io.Reader         // 请求正文内容，如果设置body，则params以及rawparam均无效
    Transport *http.Transport   // 传输协议
    Timeout   int64             // 超时时间，单位：秒
    Error     error             // 错误信息
    paramType int               // 请求方式，0-not set,1-键值对，2-数据流
}

// Cookie 定义一个基本的Cookie信息
type Cookie struct {
    Name     string
    Value    interface{}
    HttpOnly bool
}

// WithCookies option for cookies
func WithCookies(cookies []Cookie) clientOption {
    return func(c *Client) {
        if c.Cookies == nil {
            c.Cookies = make([]Cookie, 0)
        }
        c.Cookies = append(c.Cookies, cookies...)
    }
}

// WithHeaders option for headers
func WithHeaders(headers map[string]interface{}) clientOption {
    return func(c *Client) {
        if c.Headers == nil {
            c.Headers = make(map[string]string, 0)
        }
        if len(headers) == 0 {
            return
        }
        for k, v := range headers {
            c.Headers[k] = String(v)
        }
    }
}

// WithTransport option for transport
func WithTransport(t *http.Transport) clientOption {
    return func(c *Client) {
        if t == nil {
            return
        }
        c.Transport = t
    }
}

// WithTimeout option for timeout
func WithTimeout(t int64) clientOption {
    return func(c *Client) {
        if t <= 0 {
            return
        }
        c.Timeout = t
    }
}

// NewClient 创建一个http client对象
func NewClient(opts ...clientOption) *Client {
    c := &Client{
        Client:    &http.Client{},
        Headers:   make(map[string]string),
        Cookies:   make([]Cookie, 0),
        Params:    url.Values{},
        RawParam:  nil,
        Body:      nil,
        Transport: nil,
        Timeout:   30, // 默认30秒超时,可以支持全局配置
    }
    if len(opts) > 0 {
        for _, opt := range opts {
            opt(c)
        }
    }
    return c
}

// WithCookies option for cookies
func (c *Client) WithCookies(cookies []Cookie) *Client {
    if c.Cookies == nil {
        c.Cookies = make([]Cookie, 0)
    }
    c.Cookies = append(c.Cookies, cookies...)
    return c
}

// WithCookie 添加单个cookie
func (c *Client) WithCookie(cookie Cookie) *Client {
    if c.Cookies == nil {
        c.Cookies = make([]Cookie, 0)
    }
    c.Cookies = append(c.Cookies, cookie)
    return c
}

// WithHeaders 设置头信息
func (c *Client) WithHeaders(headers map[string]interface{}) *Client {
    if c.Headers == nil {
        c.Headers = make(map[string]string, 0)
    }
    if len(headers) == 0 {
        return c
    }
    for k, v := range headers {
        c.Headers[k] = String(v)
    }
    return c
}

// WithHeader 设置单个头信息
func (c *Client) WithHeader(k string, v interface{}) *Client {
    if c.Headers == nil {
        c.Headers = make(map[string]string, 0)
    }
    c.Headers[k] = String(v)
    return c
}

// WithTransport 设置Transport
func (c *Client) WithTransport(t *http.Transport) *Client {
    if t == nil {
        return c
    }
    c.Transport = t
    return c
}

// WithTimeout 设置超时时间(秒)，小于等于0的值将被忽略
func (c *Client) WithTimeout(t int64) *Client {
    if t <= 0 {
        t = 30
    }
    c.Timeout = t
    return c
}

// WithBody 设置请求正文
func (c *Client) WithBody(r io.Reader) *Client {
    c.Body = r
    return c
}

// ResetParam 重置请求参数
func (c *Client) ResetParam() *Client {
    c.Params = url.Values{}
    c.RawParam = []byte{}
    c.paramType = 0
    return c
}

// WithRawParam 设置原始请求参数
func (c *Client) WithRawParam(param []byte) *Client {
    if c.Error != nil {
        return c
    }
    c.RawParam = param
    if c.paramType > paramTypeNone && c.paramType != paramTypeStream {
        c.Error = ErrInvalidParamType
    } else {
        c.paramType = paramTypeStream
    }
    return c
}

// WithParams 设置参数，设置之前会清空原有的参数
func (c *Client) WithParams(params map[string]interface{}) *Client {
    if c.Error != nil {
        return c
    }
    if c.Params == nil {
        c.Params = url.Values{}
    }
    c.addParams(params)
    if c.paramType > paramTypeNone && c.paramType != paramTypeKV {
        c.Error = ErrInvalidParamType
    } else {
        c.paramType = paramTypeKV
    }
    return c
}

// AddParams 添加多个参数
func (c *Client) addParams(params map[string]interface{}) {
    if len(params) == 0 {
        return
    }
    for k, v := range params {
        c.addParam(k, v)
    }
}

// AddParam 添加单个参数
func (c *Client) addParam(key string, val interface{}) {
    if key == "" {
        return
    }
    rv := reflect.ValueOf(val)
    switch rv.Kind() {
    case reflect.Slice, reflect.Array:
        for i := 0; i < rv.Len(); i++ {
            c.Params.Add(key, String(rv.Index(i).Interface()))
        }
    default:
        c.Params.Add(key, String(val))
    }
}

// buildRequest 构造请求信息
func (client *Client) buildRequest(method, url string) (*http.Request, error) {
    queryString := client.buildQueryString()
    var body io.Reader
    if client.Body != nil {
        body = client.Body
    } else {
        if queryString != "" {
            if method == http.MethodGet {
                if strings.Contains(url, "?") {
                    url += "&"
                } else {
                    url += "?"
                }
                url += queryString
            } else if method == http.MethodPost {
                body = strings.NewReader(queryString)
            }
        }
    }

    // 构造请求
    request, err := http.NewRequest(method, url, body)
    if err != nil {
        return nil, err
    }
    // 设置头信息
    if len(client.Headers) > 0 {
        for k, v := range client.Headers {
            request.Header.Set(k, v)
        }
    }
    // 添加cookie信息
    if len(client.Cookies) > 0 {
        for _, c := range client.Cookies {
            request.AddCookie(&http.Cookie{Name: c.Name, Value: String(c.Value), HttpOnly: c.HttpOnly})
        }
    }
    // 设置http协议版本
    request.Proto = "HTTP/1.1"
    // 返回请求信息
    return request, nil
}

// buildQueryString 构造query string
func (c *Client) buildQueryString() string {
    if c.paramType == paramTypeStream {
        return string(c.RawParam)
    }
    // 根据设置的参数进行处理
    if len(c.Params) == 0 {
        return ""
    }
    return c.Params.Encode()
}

// initSelf 初始化http client的参数信息
func (c *Client) initSelf() *Client {
    if c.Client == nil {
        c.Client = &http.Client{}
    }
    if c.Timeout > 0 {
        c.Client.Timeout = time.Second * time.Duration(c.Timeout)
    }
    if c.Transport != nil {
        c.Client.Transport = c.Transport
    }
    return c
}

// Request 根据参数指定的方式执行请求
func (c *Client) Request(method string, url string) (*http.Response, error) {
    if c.Error != nil {
        return nil, c.Error
    }
    switch method {
    case http.MethodGet:
        return c.Get(url)
    case http.MethodPost:
        return c.Post(url)
    default:
        return nil, errors.New("request method [" + method + "] not supported")
    }
}

// Get 执行http get请求
func (c *Client) Get(url string) (*http.Response, error) {
    if c.Error != nil {
        return nil, c.Error
    }
    // 构造请求
    request, err := c.initSelf().buildRequest(http.MethodGet, url)
    if err != nil {
        return nil, err
    }
    // 执行请求
    return c.Client.Do(request)
}

// Post 执行http post请求
func (c *Client) Post(url string) (*http.Response, error) {
    if c.Error != nil {
        return nil, c.Error
    }
    // 添加默认头信息，防止post方式服务器端收不到数据
    if _, ok := c.Headers["Content-Type"]; !ok {
        c.Headers["Content-Type"] = "application/x-www-form-urlencoded"
    }
    // 构造请求
    request, err := c.initSelf().buildRequest(http.MethodPost, url)
    if err != nil {
        return nil, err
    }
    // 执行请求
    return c.Client.Do(request)
}
