package array

import (
    "github.com/whencome/goutil"
)

// Contains 检查列表是否包含某个值
func Contains[T comparable](arr []T, s T) bool {
    if len(arr) == 0 {
        return false
    }
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

// Chunk 将给定的切片拆分成指定大小size的二维切片
func Chunk[T any](arr []T, size int) [][]T {
    arraySize := len(arr)
    if arraySize <= size {
        return [][]T{arr}
    }

    ret := make([][]T, 0)
    for i := 0; ; i++ {
        begin, end := i*size, (i+1)*size

        if begin >= arraySize {
            break
        }
        if end > arraySize {
            end = arraySize
        }
        ret = append(ret, arr[begin:end])
    }
    return ret
}

// Remove 从数组中移除指定的元素
func Remove[T comparable](arr []T, e T) []T {
    newArr := make([]T, 0)
    for _, a := range arr {
        if a == e {
            continue
        }
        newArr = append(newArr, a)
    }
    return newArr
}

// RemoveBatch 从数组中移除多个元素
func RemoveBatch[T comparable](arr []T, elements ...T) []T {
    if len(elements) == 0 {
        return arr
    }
    eMaps := make(map[T]bool)
    for _, e := range elements {
        eMaps[e] = true
    }
    newArr := make([]T, 0)
    for _, a := range arr {
        if _, ok := eMaps[a]; ok {
            continue
        }
        newArr = append(newArr, a)
    }
    return newArr
}
