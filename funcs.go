package gotil

import (
    "math/rand"
    "reflect"
    "strings"
    "time"
    "unicode/utf8"

    "github.com/axgle/mahonia"
)

// 定义数据表特殊字符替换映射表
var sqlSpecialCharMaps = []map[string]string {
    {"old":`\`, "new":`\\`},
    {"old":`'`, "new":`\'`},
    {"old":`"`, "new":`\"`},
}

// EscapeSqlValue 转义数据库中的特殊字符，暂时只处理常见内容
func EscapeSqlValue(str string) string {
    // 先检查是否是utf8，不是则先转换
    if !utf8.ValidString(str) {
        utf8Encoder := mahonia.NewEncoder("UTF-8")
        str = utf8Encoder.ConvertString(str)
    }
    for _, repl := range sqlSpecialCharMaps {
        str = strings.ReplaceAll(str, repl["old"], repl["new"])
    }
    return str
}

////////////////// TYPE CONVERSION FUNCS ///////////////
func String(v interface{}) string {
    return Object(v).String()
}

func Int64(v interface{}) int64 {
    return Object(v).Int64()
}

func Int(v interface{}) int {
    return int(Object(v).Int64())
}

func Uint64(v interface{}) uint64 {
    return Object(v).Uint64()
}

func Uint(v interface{}) uint {
    return uint(Object(v).Uint64())
}

func Float64(v interface{}) float64 {
    return Object(v).Float64()
}

func Float32(v interface{}) float32 {
    return float32(Object(v).Float64())
}

func Bool(v interface{}) bool {
    return Object(v).Boolean()
}

////////////////// UTIL FUNCS ///////////////
// CopyStruct 复制结构体
func CopyStruct(src, dst interface{}) {
    dstVal := reflect.ValueOf(dst).Elem() //获取reflect.Type类型
    srcVal := reflect.ValueOf(src).Elem()   //获取reflect.Type类型
    vTypeOfT := srcVal.Type()
    for i := 0; i < srcVal.NumField(); i++ {
        // 在要修改的结构体中查询有数据结构体中相同属性的字段，有则修改其值
        name := vTypeOfT.Field(i).Name
        // 检查目标struct是否存在name
        if reflect.DeepEqual(dstVal.FieldByName(name), reflect.Value{}){
            continue
        }
        dFieldType := dstVal.FieldByName(name).Type()
        sFieldType := srcVal.FieldByName(name).Type()
        // 名字及类型一致，才能赋值
        if ok := dstVal.FieldByName(name).IsValid(); ok && sFieldType.AssignableTo(dFieldType){
            dstVal.FieldByName(name).Set(reflect.ValueOf(srcVal.Field(i).Interface()))
        }
    }
}

// 获取随机字符串
func RandString(len int) string {
    r := rand.New(rand.NewSource(time.Now().Unix()))
    bytes := make([]byte, len)
    for i := 0; i < len; i++ {
        b := r.Intn(26) + 65
        bytes[i] = byte(b)
    }
    return string(bytes)
}

// IsNil 判断给定的值是否为nil
func IsNil(i interface{}) bool {
    ret := i == nil
    // 需要进一步做判断
    if !ret {
        vi := reflect.ValueOf(i)
        kind := reflect.ValueOf(i).Kind()
        if kind == reflect.Slice ||
            kind == reflect.Map ||
            kind == reflect.Chan ||
            kind == reflect.Interface ||
            kind == reflect.Func ||
            kind == reflect.Ptr {
            return vi.IsNil()
        }
    }
    return ret
}

// 字符串截取
func SubString(str string, start, end int) string {
    rs := []rune(str)
    return string(rs[start:end])
}

// 判断是否为空
func IsEmpty(v interface{}) bool {
    if v == nil {
        return true
    }
    switch v.(type) {
    case int:
        val := v.(int)
        return val == 0
    case int8:
        val := v.(int8)
        return val == 0
    case int16:
        val := v.(int16)
        return val == 0
    case int32:
        val := v.(int32)
        return val == 0
    case int64:
        val := v.(int64)
        return val == 0
    case uint:
        val := v.(uint)
        return val == 0
    case uint8:
        val := v.(uint8)
        return val == 0
    case uint16:
        val := v.(uint16)
        return val == 0
    case uint32:
        val := v.(uint32)
        return val == 0
    case uint64:
        val := v.(uint64)
        return val == 0
    case float32:
        val := v.(float32)
        return val == 0
    case float64:
        val := v.(float64)
        return val == 0
    case string:
        val := v.(string)
        if val == "" {
            return true
        }
    case []byte:
        val := v.([]byte)
        return len(val) == 0
    case []rune:
        val := v.([]rune)
        return len(val) == 0
    case bool:
        val := v.(bool)
        return !val
    default:
        return IsNil(v)
    }
    return false
}

