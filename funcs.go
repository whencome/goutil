package gotil

import (
    "fmt"
    "math/rand"
    "reflect"
    "strconv"
    "time"
)

// //////////////// UTIL FUNCS ///////////////

// CopyStruct 复制结构体
func CopyStruct(src, dst interface{}) {
    dstVal := reflect.ValueOf(dst).Elem() // 获取reflect.Type类型
    srcVal := reflect.ValueOf(src).Elem() // 获取reflect.Type类型
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

// RandString 获取随机字符串
func RandString(len int) string {
    r := rand.New(rand.NewSource(time.Now().Unix()))
    bytes := make([]byte, len)
    for i := 0; i < len; i++ {
        b := r.Intn(26) + 65
        bytes[i] = byte(b)
    }
    return string(bytes)
}

// RandNumericString 生成一个只包含数字的随机字符串
func RandNumericString(len int) string {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    bytes := make([]byte, len)
    for i := 0; i < len; i++ {
        b := r.Intn(10) + 48
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

// IsList 判断给定的对象是否是数组或者切片
func IsList(i interface{}) bool {
    if i == nil {
        return false
    }
    value := reflect.ValueOf(i)
    if value.Kind() == reflect.Ptr {
        value = value.Elem()
    }
    if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
        return true
    }
    return false
}

// SubString 字符串截取
func SubString(str string, start, end int) string {
    rs := []rune(str)
    return string(rs[start:end])
}

// IsEmpty 判断是否为空
func IsEmpty(v interface{}) bool {
    if v == nil {
        return true
    }
    return reflect.ValueOf(v).IsZero()
}

// Cond 条件表达式，用于实现三元运算符的功能，if cond == true then trueVal else falseVal
func Cond(cond bool, trueVal interface{}, falseVal interface{}) interface{} {
    if cond {
        return trueVal
    }
    return falseVal
}

// NilCond 判断值v，如果等于nil，则返回elseVal，否则返回v本身
func NilCond(v interface{}, elseVal interface{}) interface{} {
    if v == nil {
        return elseVal
    }
    return v
}

// EmptyCond 判断值v，如果为空（零值或者nil），则返回elseVal，否则返回v本身
func EmptyCond(v interface{}, elseVal interface{}) interface{} {
    if IsEmpty(v) {
        return elseVal
    }
    return v
}

// ///////////////////// TYPE CONVERSIONS ///////////////////////////

type stringer interface {
    String() string
}

type int64er interface {
    Int64() (int64, error)
}

type uint64er interface {
    Uint64() (uint64, error)
}

type float64er interface {
    Float64() (float64, error)
}

// String 将任意值转换成string类型
// 注意：1. 当无法或者不支持转换时，此接口将返回0值；
//       2. 此接口应当只在明确数据内容只是需要类型转换时使用，不应当用于对未知值进行转换。
func String(v interface{}) string {
    var strVal = ""
    if sv, ok := v.(stringer); ok {
        strVal = sv.String()
        return strVal
    }
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
    if iv, ok := v.(int64er); ok {
        i, _ := iv.Int64()
        return i
    }
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
        n, err := strconv.ParseInt(v.(string), 10, 64)
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
    if uv, ok := v.(uint64er); ok {
        ui, _ := uv.Uint64()
        return ui
    }
    if iv, ok := v.(int64er); ok {
        ii, _ := iv.Int64()
        return uint64(ii)
    }
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
        n, err := strconv.ParseUint(v.(string), 10, 64)
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
    if fv, ok := v.(float64er); ok {
        f, _ := fv.Float64()
        return f
    }
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
        return v.(float64)
    case string:
        n, err := strconv.ParseFloat(v.(string), 64)
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
