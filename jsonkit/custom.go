package jsonkit

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// JS 安全整数范围常量
// JavaScript 的最大安全整数是 2^53 - 1 (9007199254740991)
// 最小安全整数是 -(2^53 - 1)
const (
	jsMaxSafeInteger = 9007199254740991
	jsMinSafeInteger = -9007199254740991
)

// sanitizeJSON 递归处理接口数据
// 规则：如果是 json.Number，检查是否在 JS 安全范围内。
// 范围内 -> 转为 float64 (以便 marshal 成数字)
// 范围外 -> 转为 string (以便 marshal 成字符串)
func sanitizeJSON(input interface{}) interface{} {
	switch v := input.(type) {
	case json.Number:
		// 尝试转换为 int64 以进行范围检查
		// 注意：如果数字太大超过 int64，Int64() 会报错，直接视为不安全
		i, err := v.Int64()
		if err != nil {
			// 转换失败（通常是因为超过了 int64 范围），肯定是超大数，转字符串
			return v.String()
		}

		// 检查是否在 JS 安全整数范围内
		if i >= jsMinSafeInteger && i <= jsMaxSafeInteger {
			// 安全：转换为 float64。
			// 为什么转 float64？因为 json.Marshal 遇到 float64 会输出为数字 (123)，
			// 而遇到 int64 也会输出为数字，但为了统一处理逻辑，转 float64 最稳妥，
			// 且在这个范围内 float64 能精确表示 int64。
			return float64(i)
		} else {
			// 不安全：虽然能放进 int64，但前端会丢精度，必须转字符串
			return v.String()
		}

	case map[string]interface{}:
		// 递归处理 Map
		result := make(map[string]interface{}, len(v))
		for k, val := range v {
			result[k] = sanitizeJSON(val)
		}
		return result

	case []interface{}:
		// 递归处理 Slice/Array
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = sanitizeJSON(val)
		}
		return result

	default:
		// 其他类型（string, bool, nil 等）原样返回
		return v
	}
}

// NumSafeDecode 数字安全的 json 解码
func NumSafeDecode(jsonStr string, dst interface{}) error {
	// dst 检查，dst 必须是指针类型
	if dst == nil {
		return fmt.Errorf("dst cannot be nil")
	}

	val := reflect.ValueOf(dst)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("dst must be a pointer, got %T", dst)
	}

	if val.IsNil() {
		return fmt.Errorf("dst pointer cannot be nil")
	}

	// json 解析
	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	// 【关键步骤 1】开启 UseNumber，防止解码时直接变 float64 导致精度丢失
	decoder.UseNumber()
	if err := decoder.Decode(dst); err != nil {
		return err
	}

	// 获取解码后的值并进行安全处理
	dstVal := reflect.ValueOf(dst).Elem()
	if !dstVal.CanSet() {
		// 目标不可写（如指向未导出字段等），直接返回，避免panic
		return nil
	}
	sanitizedValue := sanitizeJSON(dstVal.Interface())
	if sanitizedValue == nil {
		return nil
	}
	// 仅在类型可赋值时回写，避免sanitize过程中类型发生变化导致的panic
	sanitizedVal := reflect.ValueOf(sanitizedValue)
	if !sanitizedVal.Type().AssignableTo(dstVal.Type()) {
		return nil
	}

	// 设置回目标变量
	dstVal.Set(sanitizedVal)

	return nil
}
