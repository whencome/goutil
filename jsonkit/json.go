package jsonkit

import (
	"bytes"
	"encoding/json"
	jsoniter "github.com/json-iterator/go"
	jsoniterExtra "github.com/json-iterator/go/extra"
	"sort"
)

func init() {
	// 启动模糊模式，支持更宽松的json解析
	jsoniterExtra.RegisterFuzzyDecoders()
}

var xjson = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
	UseNumber:              true,
}.Froze()

// Marshal 将给定的对象编码为（json字符串）字节数组
func Marshal(v any) ([]byte, error) {
	return xjson.Marshal(v)
}

// MarshalIndent 带有缩进的Marshal
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return xjson.MarshalIndent(v, prefix, indent)
}

// MarshalString 将Marshal得到的字节数组转换为字符串（json）并返回
func MarshalString(v any) (string, error) {
	b, e := Marshal(v)
	return string(b), e
}

// Unmarshal 将给定的字节数组反序列化到目标对象v
func Unmarshal(data []byte, v any) error {
	return xjson.Unmarshal(data, v)
}

func UnmarshalString(data string, v any) error {
	return xjson.Unmarshal([]byte(data), v)
}

// Encode 将传入的参数编码为json字符串
// 由于此方法会忽略错误信息，使用时需要慎重
func Encode(v any) string {
	s, _ := MarshalString(v)
	return s
}

// Decode 解码json内容
func Decode(data string, v any) error {
	return UnmarshalString(data, v)
}

// MarshalSorted 将任意对象v进行json编码，并对key进行排序
func MarshalSorted(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)

	// 先编码为JSON
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}

	// 解析JSON并重新排序
	var rawData interface{}
	if err := json.Unmarshal(buf.Bytes(), &rawData); err != nil {
		return nil, err
	}

	// 递归排序所有map的key
	sortedData := sortMapKeys(rawData)

	// 重新编码（不转义）
	var resultBuf bytes.Buffer
	encoder = json.NewEncoder(&resultBuf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(sortedData); err != nil {
		return nil, err
	}

	return resultBuf.Bytes(), nil
}

// sortMapKeys 对json的key进行排序
func sortMapKeys(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		sortedMap := make(map[string]interface{})

		// 获取并排序keys
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// 按排序后的顺序插入，并递归处理值
		for _, k := range keys {
			sortedMap[k] = sortMapKeys(val[k])
		}
		return sortedMap

	case []interface{}:
		// 递归处理数组元素
		sortedArray := make([]interface{}, len(val))
		for i, item := range val {
			sortedArray[i] = sortMapKeys(item)
		}
		return sortedArray

	default:
		return v
	}
}
