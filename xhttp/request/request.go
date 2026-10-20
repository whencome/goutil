package request

import (
    "io"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/whencome/goutil"
    "github.com/whencome/goutil/jsonkit"
)

// Params 获取请求数据,暂时只支持GET & POST
func Params(c *gin.Context) map[string][]string {
    if c == nil || c.Request == nil {
        return map[string][]string{}
    }
    var d map[string][]string
    if c.Request.Method == "GET" {
        d = GetParams(c.Request)
    } else {
        d = PostParams(c.Request)
    }
    return d
}

// SimpleParams 获取简单请求数据,暂时只支持GET & POST，如果传递的是数组，则只取第一个值
func SimpleParams(c *gin.Context) goutil.M {
    if c == nil {
        return goutil.M{}
    }
    return GetSimpleParams(c.Request)
}

// GetSimpleParams 获取简单请求数据,暂时只支持GET & POST，如果传递的是数组，则只取第一个值
func GetSimpleParams(r *http.Request) goutil.M {
    m := goutil.M{}
    if r == nil {
        return m
    }
    var d map[string][]string
    if r.Method == "GET" {
        d = GetParams(r)
    } else {
        d = PostParams(r)
    }
    if len(d) > 0 {
        for k, vs := range d {
            if len(vs) > 0 {
                m[k] = vs[0]
            } else {
                m[k] = ""
            }
        }
    }
    return m
}

// GetParams 从get中获取请求数据
func GetParams(r *http.Request) map[string][]string {
    if r == nil {
        return map[string][]string{}
    }
    if err := r.ParseForm(); err != nil {
        // 解析失败时返回已解析到的部分（可能为空），避免调用方拿到 nil
        if r.Form == nil {
            return map[string][]string{}
        }
    }
    return r.Form
}

// PostParams 从post中获取请求数据
func PostParams(r *http.Request) map[string][]string {
    if r == nil {
        return map[string][]string{}
    }
    return getPostParams(r)
}

// getPostParams 根据content-type提取数据
func getPostParams(r *http.Request) map[string][]string {
    if r == nil || r.Body == nil && r.MultipartForm == nil {
        return map[string][]string{}
    }
    contentType := ContentType(r)
    if contentType == "application/x-www-form-urlencoded" {
        if err := r.ParseForm(); err != nil && r.PostForm == nil {
            return map[string][]string{}
        }
        return r.PostForm
    }
    if contentType == "multipart/form-data" {
        if err := r.ParseMultipartForm(1024 * 100); err != nil {
            // 解析失败时 MultipartForm 可能为 nil，直接访问其 Value 会 panic
            if r.MultipartForm == nil {
                return map[string][]string{}
            }
        }
        if r.MultipartForm == nil {
            return map[string][]string{}
        }
        return r.MultipartForm.Value
    }
    if contentType == "application/json" {
        if r.Body == nil {
            return map[string][]string{}
        }
        body, err := io.ReadAll(r.Body)
        if err != nil {
            return map[string][]string{}
        }
        data := make(map[string]interface{})
        err = jsonkit.Unmarshal(body, &data)
        if err != nil {
            return map[string][]string{}
        }
        result := make(map[string][]string)
        for k, v := range data {
            result[k] = []string{jsonValueToString(v)}
        }
        return result
    }
    return map[string][]string{}
}

// jsonValueToString 将json解析得到的值转换成字符串
// 注意：与goutil.String不同，这里对于数组/对象会保留json字面量形式，便于调用方二次解析
func jsonValueToString(v interface{}) string {
    switch v.(type) {
    case nil:
        return ""
    case string:
        return v.(string)
    case map[string]interface{}, []interface{}:
        return jsonkit.Encode(v)
    default:
        return goutil.String(v)
    }
}

// ContentType 获取请求的Content-Type
// 注意：若请求未设置Content-Type，则默认按 application/x-www-form-urlencoded 处理
func ContentType(r *http.Request) string {
    if r == nil || r.Header == nil {
        return "application/x-www-form-urlencoded"
    }
    contentTypes := r.Header["Content-Type"]
    if len(contentTypes) == 0 {
        return "application/x-www-form-urlencoded"
    }
    contentType := contentTypes[0]
    if strings.Contains(contentType, ";") {
        pos := strings.Index(contentType, ";")
        contentType = contentType[0:pos]
    }
    contentType = strings.TrimSpace(strings.ToLower(contentType))
    if contentType == "" {
        return "application/x-www-form-urlencoded"
    }
    return contentType
}
