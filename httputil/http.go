package httputil

import (
	"fmt"
	"github.com/whencome/gotil"
	"github.com/whencome/xlog"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	// 默认curl超时时间
	DefaultCurlTimeOut = 5
	// 请求方法
	MethodGET  = "GET"
	MethodPOST = "POST"
)

// 创建一个新的http client
func NewClient(timeout int, t *http.Transport) *HttpClient {
	if timeout <= 0 {
		timeout = DefaultCurlTimeOut
	}
	client := NewHttpClient()
	client.SetTimeout(timeout)
	client.SetTransport(t)
	// 开启调试，后期改成配置
	client.SetDebug(true)
	return client
}

// 创建一个新的http client
func NewClientWithHeaders(timeout int, headers map[string]interface{}) *HttpClient {
	if timeout <= 0 {
		timeout = DefaultCurlTimeOut
	}
	client := NewHttpClient()
	client.SetTimeout(timeout)
	client.SetHeaders(headers)
	// 开启调试，后期改成配置
	client.SetDebug(true)
	return client
}

// 执行请求
func request(client *HttpClient, method, url string, params interface{}) ([]byte, error) {
	if client == nil {
		return nil, fmt.Errorf("http client not initialized")
	}
	if method != MethodGET && method != MethodPOST {
		return nil, fmt.Errorf("request method not supported")
	}
	// 设置参数
	switch params.(type) {
	case map[string]string, map[string]interface{}:
		client.SetParams(params)
	case []byte:
		client.SetRawParam(params.([]byte))
	}

	// 执行请求
	var resp *http.Response
	var err error
	if method == MethodGET {
		resp, err = client.Get(url)
	} else {
		resp, err = client.Post(url)
	}
	// 解析并返回结果
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		xlog.Debugf("request error : %s", err)
		return nil, err
	}
	xlog.Debugf("request result: %s", string(body))
	return body, nil
}

// Request 执行请求
func Request(client *HttpClient, method, url string, params interface{}) ([]byte, error) {
	return request(client, method, url, params)
}

// 执行http get请求
func Get(url string, params map[string]string, timeout int) ([]byte, error) {
	client := NewClient(timeout, nil)
	return request(client, MethodGET, url, params)
}

// 执行http post请求
func Post(url string, params map[string]string, timeout int) ([]byte, error) {
	client := NewClient(timeout, nil)
	return request(client, MethodPOST, url, params)
}

// 执行http post请求
func PostRaw(url string, params []byte, timeout int) ([]byte, error) {
	client := NewClient(timeout, nil)
	return request(client, MethodPOST, url, params)
}

// 执行http get请求
func CustomGet(t *http.Transport, url string, params map[string]string, timeout int) ([]byte, error) {
	client := NewClient(timeout, t)
	return request(client, MethodGET, url, params)
}

// 执行http post请求
func CustomPost(t *http.Transport, url string, params map[string]string, timeout int) ([]byte, error) {
	client := NewClient(timeout, t)
	return request(client, MethodPOST, url, params)
}

// 执行http post请求, 指定传输协议
func CustomPostRaw(t *http.Transport, url string, params []byte, timeout int) ([]byte, error) {
	client := NewClient(timeout, t)
	return request(client, MethodPOST, url, params)
}

// 构造请求参数信息
func BuildQuery(data map[string]interface{}) string {
	if data == nil || len(data) == 0 {
		return ""
	}
	params := make([]string, 0)
	for k, v := range data {
		params = append(params, fmt.Sprintf("%s=%s", k, url.QueryEscape(gotil.String(v))))
	}
	return strings.Join(params, "&")
}
