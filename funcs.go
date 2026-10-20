package goutil

import (
	cryptorand "crypto/rand"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
)

// //////////////// UTIL FUNCS ///////////////

// CopyStruct 复制结构体
// 说明：将src中与dst同名的字段值复制到dst中，src和dst均支持结构体或结构体指针；
// 若src或dst为nil、非结构体、或目标字段不可写，则跳过对应字段而不会panic。
func CopyStruct(src, dst interface{}) {
	if src == nil || dst == nil {
		return
	}
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)
	// 目标必须是一个可写的指针
	if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		return
	}
	if srcVal.Kind() == reflect.Ptr {
		if srcVal.IsNil() {
			return
		}
		srcVal = srcVal.Elem()
	}
	dstVal = dstVal.Elem()
	if srcVal.Kind() != reflect.Struct || dstVal.Kind() != reflect.Struct {
		return
	}
	srcType := srcVal.Type()
	for i := 0; i < srcVal.NumField(); i++ {
		// 在要修改的结构体中查询有数据结构体中相同属性的字段，有则修改其值
		name := srcType.Field(i).Name
		dField := dstVal.FieldByName(name)
		// 目标结构体中不存在该字段，或字段不可写（如未导出字段）
		if !dField.IsValid() || !dField.CanSet() {
			continue
		}
		sField := srcVal.Field(i)
		if !sField.CanInterface() {
			continue
		}
		// 名字及类型一致，才能赋值
		if sField.Type().AssignableTo(dField.Type()) {
			dField.Set(sField)
		}
	}
}

// DeepCopy 深度复制一个值，处理嵌套的map、slice、指针和结构体
// 返回一个新的独立副本，修改副本不会影响原始值
// 注意：map类型统一返回 map[string]interface{}（key会被转换为string），此行为为兼容历史版本而保留
func DeepCopy(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	val := reflect.ValueOf(v)
	if !val.IsValid() {
		return nil
	}

	// 处理指针类型
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		// 递归复制指针指向的值
		elemCopy := DeepCopy(val.Elem().Interface())
		if elemCopy == nil {
			// 指向nil的指针，返回同类型的nil指针
			return reflect.Zero(val.Type()).Interface()
		}
		// 创建新的指针指向复制后的值
		newPtr := reflect.New(reflect.TypeOf(elemCopy))
		newPtr.Elem().Set(reflect.ValueOf(elemCopy))
		return newPtr.Interface()
	}

	switch val.Kind() {
	case reflect.Map:
		// 创建新的map
		if val.IsNil() {
			return map[string]interface{}(nil)
		}
		newMap := make(map[string]interface{}, val.Len())
		for _, key := range val.MapKeys() {
			strKey := String(key.Interface())
			newMap[strKey] = DeepCopy(val.MapIndex(key).Interface())
		}
		return newMap
	case reflect.Slice:
		if val.IsNil() {
			return reflect.Zero(val.Type()).Interface()
		}
		// 创建新的slice，保持原始类型
		newSlice := reflect.MakeSlice(val.Type(), val.Len(), val.Len())
		for i := 0; i < val.Len(); i++ {
			elemCopy := DeepCopy(val.Index(i).Interface())
			if elemCopy == nil {
				continue
			}
			newSlice.Index(i).Set(reflect.ValueOf(elemCopy))
		}
		return newSlice.Interface()
	case reflect.Array:
		// 数组类型不能使用MakeSlice，需要创建同类型的新数组
		newArr := reflect.New(val.Type()).Elem()
		for i := 0; i < val.Len(); i++ {
			elemCopy := DeepCopy(val.Index(i).Interface())
			if elemCopy == nil {
				continue
			}
			newArr.Index(i).Set(reflect.ValueOf(elemCopy))
		}
		return newArr.Interface()
	case reflect.Struct:
		// 对于结构体，创建一个新的实例并复制所有字段
		newStruct := reflect.New(val.Type()).Elem()
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if field.CanInterface() && newStruct.Field(i).CanSet() {
				fieldCopy := DeepCopy(field.Interface())
				if fieldCopy == nil {
					continue
				}
				newStruct.Field(i).Set(reflect.ValueOf(fieldCopy))
			}
		}
		return newStruct.Interface()
	default:
		// 对于基本类型和其他不可变类型，直接返回
		return v
	}
}

// RandString 获取随机字符串（仅包含大写字母）
// 说明：使用全局并发安全的随机源，同一秒内多次调用也会返回不同的结果；
// 此方法不适用于生成密钥、令牌等安全敏感场景，安全敏感场景请使用 RandStringSecure
func RandString(len int) string {
	if len <= 0 {
		return ""
	}
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte('A' + rand.Intn(26))
	}
	return string(bytes)
}

// RandNumericString 生成一个只包含数字的随机字符串
// 说明：使用全局并发安全的随机源，同一秒内多次调用也会返回不同的结果
func RandNumericString(len int) string {
	if len <= 0 {
		return ""
	}
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte('0' + rand.Intn(10))
	}
	return string(bytes)
}

// RandStringSecure 使用加密安全的随机源生成一个随机字符串
// 字符集为大写字母、小写字母与数字，适用于生成令牌、随机密钥等安全敏感场景
func RandStringSecure(len int) (string, error) {
	if len <= 0 {
		return "", nil
	}
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	// 注意：参数名len与内建函数len同名，此处不能直接使用len(charset)
	const charsetSize = 62
	buf := make([]byte, len)
	if _, err := cryptorand.Read(buf); err != nil {
		return "", err
	}
	for i := range buf {
		buf[i] = charset[int(buf[i])%charsetSize]
	}
	return string(buf), nil
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

// IsMap 判断给定的值是否是一个map
func IsMap(v interface{}) bool {
	if v == nil {
		return false
	}
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	return value.Kind() == reflect.Map
}

// SubString 字符串截取
// 说明：按rune（字符）截取，start与end均支持负数（表示从末尾计算）；
// 当start或end越界时会自动收敛到有效范围，start大于等于end时返回空字符串，
// 因此本方法不会因为参数越界而panic
func SubString(str string, start, end int) string {
	rs := []rune(str)
	length := len(rs)
	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length + end
	}
	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start >= length || start >= end {
		return ""
	}
	if end < 0 {
		return ""
	}
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
//  2. 此接口应当只在明确数据内容只是需要类型转换时使用，不应当用于对未知值进行转换。
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
//  2. 此接口应当只在明确数据内容只是需要类型转换时使用，不应当用于对未知值进行转换。
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
		// 兼容带空白字符的字符串，如 " 123 "
		n, err := strconv.ParseInt(strings.TrimSpace(v.(string)), 10, 64)
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
}

func Int32(v interface{}) int32 {
	return int32(Int64(v))
}

func Int(v interface{}) int {
	return int(Int64(v))
}

// Uint64 将任意值转换成uint64类型
// 注意：1. 当无法或者不支持转换时，此接口将返回0值；
//  2. 此接口应当只在明确数据内容只是需要类型转换时使用，不应当用于对未知值进行转换。
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
		// 兼容带空白字符的字符串，如 " 123 "
		n, err := strconv.ParseUint(strings.TrimSpace(v.(string)), 10, 64)
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
}

func Uint32(v interface{}) uint32 {
	return uint32(Uint64(v))
}

func Uint(v interface{}) uint {
	return uint(Uint64(v))
}

// Float64 将任意值转换成float64类型
// 注意：1. 当无法或者不支持转换时，此接口将返回0值；
//  2. 此接口应当只在明确数据内容只是需要类型转换时使用，不应当用于对未知值进行转换。
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
		// 兼容带空白字符的字符串，如 " 1.23 "
		n, err := strconv.ParseFloat(strings.TrimSpace(v.(string)), 64)
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
}

func Float32(v interface{}) float32 {
	return float32(Float64(v))
}

// Bool 将任意值转换成bool类型
// 注意：1. 当无法或者不支持转换时，此接口将返回false；
//  2. 此接口应当只在明确数据内容只是需要类型转换时使用，不应当用于对未知值进行转换；
//  3. 字符串会先按 strconv.ParseBool 解析（支持 "true"/"false"/"T"/"F"/"1"/"0" 等），
//     解析失败后再按数字处理，因此 Bool("true") 与 Bool("1") 均返回true。
func Bool(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	if s, ok := v.(string); ok {
		s = strings.TrimSpace(s)
		if b, err := strconv.ParseBool(s); err == nil {
			return b
		}
		// 兼容 "1.0" 之类的数字字符串
		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return f != 0
		}
		return false
	}
	iv := Uint64(v)
	if iv > 0 {
		return true
	}
	return false
}
