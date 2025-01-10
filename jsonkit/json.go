package jsonkit

import (
    jsoniter "github.com/json-iterator/go"
    jsoniterExtra "github.com/json-iterator/go/extra"
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

func Marshal(v any) ([]byte, error) {
    return xjson.Marshal(v)
}

func MarshalString(v any) (string, error) {
    b, e := Marshal(v)
    return string(b), e
}

func Unmarshal(data []byte, v any) error {
    return xjson.Unmarshal(data, v)
}

func UnmarshalString(data string, v any) error {
    return xjson.Unmarshal([]byte(data), v)
}
