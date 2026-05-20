package request

import (
    "goutil"
    "goutil/jsonkit"
    "io"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
)

// Params 获取请求数据,暂时只支持GET & POST
func Params(c *gin.Context) map[string][]string {
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
    return GetSimpleParams(c.Request)
}

// GetSimpleParams 获取简单请求数据,暂时只支持GET & POST，如果传递的是数组，则只取第一个值
func GetSimpleParams(r *http.Request) goutil.M {
    var d map[string][]string
    if r.Method == "GET" {
        d = GetParams(r)
    } else {
        d = PostParams(r)
    }
    m := goutil.M{}
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
    r.ParseForm()
    return r.Form
}

// PostParams 从post中获取请求数据
func PostParams(r *http.Request) map[string][]string {
    return getPostParams(r)
}

// getPostParams 根据content-type提取数据
func getPostParams(r *http.Request) map[string][]string {
    contentType := ContentType(r)
    if contentType == "application/x-www-form-urlencoded" {
        r.ParseForm()
        return r.PostForm
    }
    if contentType == "multipart/form-data" {
        r.ParseMultipartForm(1024 * 100)
        return r.MultipartForm.Value
    }
    if contentType == "application/json" {
        body, err := io.ReadAll(r.Body)
        if err != nil {
            return nil
        }
        data := make(map[string]interface{})
        err = jsonkit.Unmarshal(body, &data)
        if err != nil {
            return nil
        }
        result := make(map[string][]string)
        for k, v := range data {
            result[k] = []string{goutil.String(v)}
        }
        return result
    }
    return nil
}

// ContentType 获取请求的Content-Type
func ContentType(r *http.Request) string {
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
    return contentType
}
