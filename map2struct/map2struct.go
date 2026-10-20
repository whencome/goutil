package map2struct

import (
    "errors"
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
    err          error             // 初始化过程中的错误信息
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
// 说明：历史实现在obj不是结构体（或结构体指针）时会panic，此处改为记录错误并返回
func (c *converter) init() {
    if c.Object == nil {
        c.err = errors.New("map2struct: object is nil")
        return
    }
    // 获取tag中的内容
    rt := reflect.TypeOf(c.Object)
    if rt.Kind() == reflect.Ptr {
        rt = rt.Elem()
    }
    rv := reflect.ValueOf(c.Object)
    if rv.Kind() == reflect.Ptr {
        if rv.IsNil() {
            c.err = errors.New("map2struct: object is a nil pointer")
            return
        }
        rv = rv.Elem()
    }
    if rt.Kind() != reflect.Struct || rv.Kind() != reflect.Struct {
        c.err = fmt.Errorf("map2struct: object must be a struct or a struct pointer, got %T", c.Object)
        return
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
        // 未导出字段无法通过反射读取值，直接跳过，避免panic
        // 注意：这里不能要求字段可写，因为ToMap允许传入非指针的结构体值
        reflectField := rv.Field(i)
        if !reflectField.CanInterface() {
            continue
        }
        // 非法字段（如匿名结构体中的冲突字段）跳过
        if !reflectField.IsValid() {
            continue
        }
        c.PropFieldMap[propName] = fieldName
        c.FieldPropMap[fieldName] = propName
        // 获取对应字段的值
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
    if c.err != nil {
        return c.err
    }
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
        if !reflectField.IsValid() || !reflectField.CanSet() {
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
        case reflect.Float64, reflect.Float32:
            reflectField.SetFloat(goutil.Float64(v))
        default: // 其他类型暂不支持
            break
        }
    }
    return nil
}

// getMapData 将给定的参数转换成指定的map格式，如果参数不是支持的类型将报错
func getMapData(data interface{}) (goutil.M, error) {
    if data == nil {
        return nil, errors.New("map2struct: data is nil")
    }
    if dm, ok := data.(goutil.M); ok {
        return dm, nil
    }
    retData := goutil.M{}
    value := reflect.ValueOf(data)
    // 支持传入map指针
    if value.Kind() == reflect.Ptr {
        if value.IsNil() {
            return nil, fmt.Errorf("map2struct: data is a nil pointer")
        }
        value = value.Elem()
    }
    if value.Kind() == reflect.Map {
        for _, k := range value.MapKeys() {
            // 使用通用的转换方法，兼容非string类型的key
            retData[goutil.String(k.Interface())] = value.MapIndex(k).Interface()
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
