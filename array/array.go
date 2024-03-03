package array

import (
    "github.com/whencome/goutil"
)

// Contains 检查列表是否包含某个值
func Contains[T comparable](arr []T, s T) bool {
    for _, v := range arr {
        if v == s {
            return true
        }
    }
    return false
}

// Unique 对给定的列表进行去重处理
func Unique[T comparable](arr []T) []T {
    ret := make([]T, 0)
    tmp := make(map[T]struct{})
    if len(arr) == 0 {
        return ret
    }
    for _, v := range arr {
        if _, ok := tmp[v]; ok {
            continue
        }
        tmp[v] = struct{}{}
        ret = append(ret, v)
    }
    return ret
}

// Filter 过滤掉列表中的空值
func Filter[T comparable](arr []T) []T {
    ret := make([]T, 0)
    if len(arr) == 0 {
        return ret
    }
    for _, v := range arr {
        if goutil.IsEmpty(v) {
            continue
        }
        ret = append(ret, v)
    }
    return ret
}

// UniqueAndFilter 对给定的列表进行去重处理，并去除空值
func UniqueAndFilter[T comparable](arr []T) []T {
    ret := make([]T, 0)
    tmp := make(map[T]struct{})
    if len(arr) == 0 {
        return ret
    }
    for _, v := range arr {
        if _, ok := tmp[v]; ok {
            continue
        }
        if goutil.IsEmpty(v) {
            continue
        }
        tmp[v] = struct{}{}
        ret = append(ret, v)
    }
    return ret
}
