package jsonkit

import (
	"strconv"
	"strings"
)

// parsePathKey 解析key，用于接续如“key[2]”格式的key
// key[2]表示字段为"key"，且该字段是一个列表， 且取索引为2的值
// 多级路径格式如：p1.p2[0].p3.p4[2]
func parsePathKey(s string) (string, int) {
	index := -1 // 默认值为-1

	// 查找最后一个'['的位置
	openBracket := strings.LastIndex(s, "[")
	if openBracket == -1 {
		// 没有找到'['，返回整个字符串和-1
		return s, index
	}

	// 检查']'是否是最后一个字符
	if len(s) == 0 || s[len(s)-1] != ']' {
		// ']'不是最后一个字符，返回整个字符串和-1
		return s, index
	}

	// 提取索引部分
	indexStr := s[openBracket+1 : len(s)-1]
	idx, err := strconv.Atoi(indexStr)
	if err != nil {
		// 索引不是数字，返回名称部分和-1
		return s[:openBracket], index
	}

	// 提取名称部分
	name := s[:openBracket]
	index = idx

	return name, index
}

// splitPath 分割路径
func splitPath(path string) []string {
    paths := strings.Split(path, "/")
    ret := make([]string, 0, len(paths))
    // 去除路径中的空字符串
    for _, p := range paths {
        p = strings.TrimSpace(p)
        if p != "" {
            ret = append(ret, p)
        }
    }
    return ret
}

// Get 从json字符串js中读取指定路径path（以/作为路径分割）的值
// 如果指定路径不存在，返回nil
func Get(js, path string) interface{} {
	m := make(map[string]interface{})
	err := Unmarshal([]byte(js), &m)
	if err != nil || len(m) == 0 {
		return nil
	}
	paths := splitPath(path)
	size := len(paths)
	src := m
	for i, p := range paths {
		if len(src) == 0 {
			return nil
		}
		key, idx := parsePathKey(p)
		if idx >= 0 {
			s, ok := src[key]
			if !ok {
				return nil
			}
			sv, ok := s.([]interface{})
			if !ok || len(sv) < idx {
				return nil
			}
			val := sv[idx]
			if i == size-1 {
				return val
			}
			src, ok = val.(map[string]interface{})
			if !ok {
				return nil
			}
			continue
		}
		s, ok := src[key]
		if !ok {
			return nil
		}
		if i == size-1 {
			return s
		}
		src, ok = s.(map[string]interface{})
		if !ok {
			return nil
		}
		continue
	}
	return nil
}
