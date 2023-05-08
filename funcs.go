package gotil

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/axgle/mahonia"
)

// 定义数据表特殊字符替换映射表
var sqlSpecialCharMaps = []map[string]string{
	{"old": `\`, "new": `\\`},
	{"old": `'`, "new": `\'`},
	{"old": `"`, "new": `\"`},
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

////////////////// UTIL FUNCS ///////////////
// CopyStruct 复制结构体
func CopyStruct(src, dst interface{}) {
	dstVal := reflect.ValueOf(dst).Elem() //获取reflect.Type类型
	srcVal := reflect.ValueOf(src).Elem() //获取reflect.Type类型
	vTypeOfT := srcVal.Type()
	for i := 0; i < srcVal.NumField(); i++ {
		// 在要修改的结构体中查询有数据结构体中相同属性的字段，有则修改其值
		name := vTypeOfT.Field(i).Name
		// 检查目标struct是否存在name
		if reflect.DeepEqual(dstVal.FieldByName(name), reflect.Value{}) {
			continue
		}
		dFieldType := dstVal.FieldByName(name).Type()
		sFieldType := srcVal.FieldByName(name).Type()
		// 名字及类型一致，才能赋值
		if ok := dstVal.FieldByName(name).IsValid(); ok && sFieldType.AssignableTo(dFieldType) {
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

/////////////////////// TYPE CONVERSIONS ///////////////////////////

// String 将任意值转换成string类型
// 注意：1. 当无法或者不支持转换时，此接口将返回0值；
//       2. 此接口应当只在明确数据内容只是需要类型转换时使用，不应当用于对未知值进行转换。
func String(v interface{}) string {
	var strVal = ""
	switch v.(type) {
	case int, int8, int16, int32, int64:
		n := Int64(v)
		strVal = strconv.FormatInt(n, 10)
	case uint, uint8, uint16, uint32, uint64:
		n := Uint64(v)
		strVal = strconv.FormatUint(n, 10)
	case float32:
		strVal = strconv.FormatFloat(float64(v.(float32)), 'f', -1, 64)
	case float64:
		strVal = strconv.FormatFloat(v.(float64), 'f', -1, 64)
	case string:
		strVal = v.(string)
	case []byte:
		strVal = string(v.([]byte))
	case []rune:
		strVal = string(v.([]rune))
	case bool:
		strVal = strconv.FormatBool(v.(bool))
	default:
		if v == nil {
			strVal = ""
		} else {
			strVal = fmt.Sprint(v)
		}
	}
	return strVal
}

// Int64 将任意值转换成int64类型
// 注意：1. 当无法或者不支持转换时，此接口将返回0值；
//       2. 此接口应当只在明确数据内容只是需要类型转换时使用，不应当用于对未知值进行转换。
func Int64(v interface{}) int64 {
	switch v.(type) {
	case int:
		return int64(v.(int))
	case int8:
		return int64(v.(int8))
	case int16:
		return int64(v.(int16))
	case int32:
		return int64(v.(int32))
	case int64:
		return v.(int64)
	case uint:
		return int64(v.(uint))
	case uint8:
		return int64(v.(uint8))
	case uint16:
		return int64(v.(uint16))
	case uint32:
		return int64(v.(uint32))
	case uint64:
		return int64(v.(uint64))
	case float32:
		return int64(v.(float32))
	case float64:
		return int64(v.(float64))
	case string:
		n, err := strconv.ParseInt(string(v.(string)), 10, 64)
		if err != nil {
			return 0
		}
		return n
	case bool:
		if v.(bool) {
			return 1
		}
		return 0
	default:
		return 0
	}
	return 0
}

func Int32(v interface{}) int32 {
	return int32(Int64(v))
}

func Int(v interface{}) int {
	return int(Int64(v))
}

// Uint64 将任意值转换成uint64类型
// 注意：1. 当无法或者不支持转换时，此接口将返回0值；
//       2. 此接口应当只在明确数据内容只是需要类型转换时使用，不应当用于对未知值进行转换。
func Uint64(v interface{}) uint64 {
	switch v.(type) {
	case int:
		return uint64(v.(int))
	case int8:
		return uint64(v.(int8))
	case int16:
		return uint64(v.(int16))
	case int32:
		return uint64(v.(int32))
	case int64:
		return uint64(v.(int64))
	case uint:
		return uint64(v.(uint))
	case uint8:
		return uint64(v.(uint8))
	case uint16:
		return uint64(v.(uint16))
	case uint32:
		return uint64(v.(uint32))
	case uint64:
		return v.(uint64)
	case float32:
		return uint64(v.(float32))
	case float64:
		return uint64(v.(float64))
	case string:
		n, err := strconv.ParseUint(string(v.(string)), 10, 64)
		if err != nil {
			return 0
		}
		return n
	case bool:
		if v.(bool) {
			return 1
		}
		return 0
	default:
		return 0
	}
	return 0
}

func Uint32(v interface{}) uint32 {
	return uint32(Uint64(v))
}

func Uint(v interface{}) uint {
	return uint(Uint64(v))
}

// Float64 将任意值转换成float64类型
// 注意：1. 当无法或者不支持转换时，此接口将返回0值；
//       2. 此接口应当只在明确数据内容只是需要类型转换时使用，不应当用于对未知值进行转换。
func Float64(v interface{}) float64 {
	switch v.(type) {
	case int:
		return float64(v.(int))
	case int8:
		return float64(v.(int8))
	case int16:
		return float64(v.(int16))
	case int32:
		return float64(v.(int32))
	case int64:
		return float64(v.(int64))
	case uint:
		return float64(v.(uint))
	case uint8:
		return float64(v.(uint8))
	case uint16:
		return float64(v.(uint16))
	case uint32:
		return float64(v.(uint32))
	case uint64:
		return float64(v.(uint64))
	case float32:
		return float64(v.(float32))
	case float64:
		return float64(v.(float64))
	case string:
		n, err := strconv.ParseFloat(string(v.(string)), 64)
		if err != nil {
			return 0
		}
		return n
	case bool:
		if v.(bool) {
			return 1
		}
		return 0
	default:
		return 0
	}
	return 0
}

func Float32(v interface{}) float32 {
	return float32(Float64(v))
}

// Bool 将任意值转换成bool类型
// 注意：1. 当无法或者不支持转换时，此接口将返回0值；
//       2. 此接口应当只在明确数据内容只是需要类型转换时使用，不应当用于对未知值进行转换。
func Bool(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	iv := Uint64(v)
	if iv > 0 {
		return true
	}
	return false
}
