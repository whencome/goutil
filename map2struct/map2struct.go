package map2struct

import (
    "fmt"
    "reflect"

    "github.com/whencome/goutil"
)

// 定义转换器
type converter struct {
    Object       interface{}       // 转换目标struct
    Tag          string            // 标签名称
    PropFieldMap map[string]string // 属性与字段(tag定义)映射关系
    FieldPropMap map[string]string // 字段与属性映射关系
    PropValues   goutil.M          // 属性与值的映射关系
}

// 创建一个转换器
func newConverter(obj interface{}, tag string) *converter {
    c := &converter{
        Object:       obj,
        Tag:          tag,
        PropFieldMap: make(map[string]string),
        FieldPropMap: make(map[string]string),
        PropValues:   goutil.M{},
    }
    c.init()
    return c
}

// 初始化转换器，提取字段信息
func (c *converter) init() {
    // 获取tag中的内容
    rt := reflect.TypeOf(c.Object)
    if rt.Kind() == reflect.Ptr {
        rt = rt.Elem()
    }
    rv := reflect.ValueOf(c.Object)
    if rv.Kind() == reflect.Ptr {
        rv = rv.Elem()
    }
    // 获取字段数量
    fieldsNum := rt.NumField()
    for i := 0; i < fieldsNum; i++ {
        field := rt.Field(i)
        propName := field.Name
        fieldName := ""
        // 如果没有指定tag，则使用属性本身作为字段名
        if c.Tag == "" {
            fieldName = propName
        } else {
            fieldName = field.Tag.Get(c.Tag)
            if fieldName == "" || fieldName == "-" {
                continue
            }
        }
        c.PropFieldMap[propName] = fieldName
        c.FieldPropMap[fieldName] = propName
        // 获取对应字段的值
        reflectField := rv.Field(i)
        c.PropValues[propName] = reflectField.Interface()
    }
}

// ToMap 将对象转换为map
func (c *converter) ToMap() goutil.M {
    // 创建m对象
    m := goutil.M{}
    for propName, value := range c.PropValues {
        fieldName, ok := c.PropFieldMap[propName]
        if !ok {
            continue
        }
        m[fieldName] = value
    }
    return m
}

// ToStruct 将map赋值到struct
func (c *converter) ToStruct(v interface{}) error {
    // 将数据转换成map
    mapData, err := getMapData(v)
    if err != nil {
        return err
    }
    if len(c.PropFieldMap) == 0 {
        return nil
    }
    // 创建对象并进行转换
    rv := reflect.ValueOf(c.Object)
    if rv.Kind() == reflect.Ptr {
        rv = rv.Elem()
    }
    for fieldName, propName := range c.FieldPropMap {
        // 检查在map中是否存在此值
        v, ok := mapData[fieldName]
        if !ok {
            continue
        }
        // 设置值
        reflectField := rv.FieldByName(propName)
        if !reflectField.CanSet() {
            continue
        }
        propTypeKind := reflectField.Kind()
        switch propTypeKind {
        case reflect.String:
            reflectField.SetString(goutil.String(v))
        case reflect.Bool:
            reflectField.SetBool(goutil.Bool(v))
        case reflect.Int64, reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8:
            reflectField.SetInt(goutil.Int64(v))
        case reflect.Uint64, reflect.Uint, reflect.Uint32, reflect.Uint16, reflect.Uint8:
            reflectField.SetUint(goutil.Uint64(v))
        case reflect.Float64:
            reflectField.SetFloat(goutil.Float64(v))
        default: // 其他类型暂不支持
            break
        }
    }
    return nil
}

// getMapData 将给定的参数转换成指定的map格式，如果参数不是支持的类型将报错
func getMapData(data interface{}) (goutil.M, error) {
    if dm, ok := data.(goutil.M); ok {
        return dm, nil
    }
    retData := goutil.M{}
    value := reflect.ValueOf(data)
    if value.Kind() == reflect.Map {
        for _, k := range value.MapKeys() {
            retData[k.String()] = value.MapIndex(k).Interface()
        }
    } else {
        return nil, fmt.Errorf("unsupported data type %T of %#v", data, data)
    }
    return retData, nil
}

// ToStructByTag 将map数据（data）填充到给定的结构体（obj）中，根据指定的tag解析与转换
func ToStructByTag(dstObj interface{}, mapData interface{}, tag string) error {
    c := newConverter(dstObj, tag)
    return c.ToStruct(mapData)
}

// ToStruct 将map数据（data）填充到给定的结构体（obj）中，根据指定的tag解析与转换
func ToStruct(dstObj interface{}, mapData interface{}) error {
    c := newConverter(dstObj, "map2struct")
    return c.ToStruct(mapData)
}

// ToMapByTag 将struct数据解析为map
func ToMapByTag(dstObj interface{}, tag string) goutil.M {
    c := newConverter(dstObj, tag)
    return c.ToMap()
}

// ToMap 将struct数据解析为map
func ToMap(dstObj interface{}) goutil.M {
    c := newConverter(dstObj, "map2struct")
    return c.ToMap()
}
