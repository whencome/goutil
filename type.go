package gotil

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/url"
    "reflect"
    "sort"
    "strings"
    "sync"
)

// ///////////////////// COMMON FUNCS ////////////////////////

// MVal 将给定V转换成一个M对象，如果不支持转换，则直接返回空的M值
func MVal(v interface{}) M {
    if v == nil {
        return M{}
    }
    // 检查是否可以直接进行类型转换
    if mv, ok := v.(M); ok {
        return mv
    }
    // 使用反射进行类型检查
    mv := M{}
    value := reflect.ValueOf(v)
    if value.Kind() == reflect.Map {
        for _, k := range value.MapKeys() {
            mv[k.String()] = value.MapIndex(k).Interface()
        }
    }
    return mv
}

// SVal 将v转换为一个Slice值，如果不支持转换，则返回空，使用者需要预先确认值类型
func SVal(v interface{}) S {
    sv := S{}
    if v == nil {
        return sv
    }
    value := reflect.ValueOf(v)
    if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
        for i := 0; i < value.Len(); i++ {
            sv = append(sv, value.Index(i).Interface())
        }
    }
    return sv
}

// XSVal 将v转换为一个增强的Slice对象，如果不支持转换，则返回空，使用者需要预先确认值类型
func XSVal(v interface{}, f func(interface{}) string) *XS {
    if xs, ok := v.(*XS); ok {
        return xs
    }

    xs := NewXS(f)
    value := reflect.ValueOf(v)
    switch value.Kind() {
    case reflect.Slice:
        for i := 0; i < value.Len(); i++ {
            xs.Add(value.Index(i).Interface())
        }
    }

    // Return empty S if conversion is not possible
    return xs
}

// ////////////////// TYPE OF KVPairs  ///////////////////////

// KVPair 参数键值对
type KVPair struct {
    K, V string
}

// KVPairs 参数键值对列表
type KVPairs []KVPair

// Less 实现排序接口
func (p KVPairs) Less(i, j int) bool {
    return p[i].K < p[j].K
}

func (p KVPairs) Swap(i, j int) {
    p[i], p[j] = p[j], p[i]
}

func (p KVPairs) Len() int {
    return len(p)
}

// Sort 排序
func (p KVPairs) Sort() {
    sort.Sort(p)
}

// Reverse 倒序
func (p KVPairs) Reverse() {
    sort.Reverse(p)
}

// Add 添加键值对
func (p KVPairs) Add(k, v string) {
    p = append(p, KVPair{K: k, V: v})
}

// FilterEmptyValueAndKeys 去掉列表中值为空以及指定字段的值
func (p KVPairs) FilterEmptyValueAndKeys(keys ...string) KVPairs {
    newKV := KVPairs{}
    sKeys := SVal(keys)
    for _, kv := range p {
        if kv.V != "" && !sKeys.Contains(kv.K) {
            newKV = append(newKV, kv)
        }
    }
    return newKV
}

// String 将所有元素按照“参数=参数值”的模式用“&”字符拼接成字符串
func (p KVPairs) String() string {
    strs := make([]string, 0)
    for _, kv := range p {
        strs = append(strs, kv.K+"="+kv.V)
    }
    return strings.Join(strs, "&")
}

// QuotedString 将所有元素按照“参数="参数值"”的模式用“&”字符拼接成字符串
func (p KVPairs) QuotedString() string {
    var strs []string
    for _, kv := range p {
        strs = append(strs, kv.K+"=\""+kv.V+"\"")
    }
    return strings.Join(strs, "&")
}

// EscapedString 将所有元素按照“参数=urlencode(参数值)”的模式用“&”字符拼接成字符串
func (p KVPairs) EscapedString() string {
    var strs []string
    for _, kv := range p {
        strs = append(strs, kv.K+"="+url.QueryEscape(kv.V))
    }
    return strings.Join(strs, "&")
}

// QuotedAndEscapedString 自定义值的处理方式
func (p KVPairs) QuotedAndEscapedString() string {
    var strs []string
    for _, kv := range p {
        v := "\"" + url.QueryEscape(kv.V) + "\""
        strs = append(strs, kv.K+"="+v)
    }
    return strings.Join(strs, "&")
}

// ToXML 将键值对转换成xml
func (p KVPairs) ToXML() string {
    buf := bytes.Buffer{}
    buf.WriteString("<xml>")
    for _, pair := range p {
        buf.WriteString(fmt.Sprintf("<%s><![CDATA[%s]]></%s>", pair.K, pair.V, pair.K))
    }
    buf.WriteString("</xml>")
    return buf.String()
}

// ToMap 将键值对转换为Map
func (p KVPairs) ToMap() map[string]string {
    m := make(map[string]string)
    for _, pair := range p {
        m[pair.K] = pair.V
    }
    return m
}

// Map2KVPairs 将map转换为键值对列表
func Map2KVPairs(params map[string]string) KVPairs {
    p := KVPairs{}
    for k, v := range params {
        p = append(p, KVPair{K: k, V: v})
    }
    return p
}

// ///////////////////// TYPE OF M ///////////////////////////

// M 常用的map类型
type M map[string]interface{}

// Pairs 将M转换为KVPairs
func (m M) Pairs() KVPairs {
    pairs := KVPairs{}
    for k, v := range m {
        pair := KVPair{
            K: k,
            V: String(v),
        }
        pairs = append(pairs, pair)
    }
    return pairs
}

// Json 获取json格式数据
func (m M) Json() string {
    v, _ := json.Marshal(m)
    return string(v)
}

// Contains 检查是否包含指定key
func (m M) Contains(k string) bool {
    if _, ok := m[k]; ok {
        return true
    }
    return false
}

// Size 获取参数数量
func (m M) Size() int {
    return len(m)
}

// Get 获取配置项k的值
func (m M) Get(k string) (interface{}, bool) {
    if v, ok := m[k]; ok {
        return v, true
    }
    return nil, false
}

// DefaultGet 获取配置项k的值，如果不存在则返回默认值
func (m M) DefaultGet(k string, dv interface{}) interface{} {
    if v, ok := m[k]; ok {
        return v
    }
    return dv
}

// DefaultGetString 获取指定K的值，如果K不存在，则返回默认值
func (m M) DefaultGetString(k, s string) string {
    v, ok := m.Get(k)
    if !ok {
        return s
    }
    return String(v)
}

func (m M) GetString(k string) string {
    return m.DefaultGetString(k, "")
}

func (m M) GetJson(k string) string {
    v, ok := m.Get(k)
    if !ok {
        return ""
    }
    b, e := json.Marshal(v)
    if e != nil {
        return ""
    }
    return string(b)
}

func (m M) DefaultGetBool(k string, b bool) bool {
    v, ok := m.Get(k)
    if !ok {
        return b
    }
    return Bool(v)
}

func (m M) GetBool(k string) bool {
    return m.DefaultGetBool(k, false)
}

func (m M) DefaultGetInt64(k string, i int64) int64 {
    v, ok := m.Get(k)
    if !ok {
        return i
    }
    return Int64(v)
}

func (m M) GetInt64(k string) int64 {
    return m.DefaultGetInt64(k, 0)
}

func (m M) DefaultGetInt(k string, i int) int {
    return int(m.DefaultGetInt64(k, int64(i)))
}

func (m M) GetInt(k string) int {
    return m.DefaultGetInt(k, 0)
}

func (m M) DefaultGetUint64(k string, i uint64) uint64 {
    v, ok := m.Get(k)
    if !ok {
        return i
    }
    return Uint64(v)
}

func (m M) GetUint64(k string) uint64 {
    return m.DefaultGetUint64(k, 0)
}

func (m M) DefaultGetUint(k string, i uint) uint {
    return uint(m.DefaultGetUint64(k, uint64(i)))
}

func (m M) GetUint(k string) uint {
    return m.DefaultGetUint(k, 0)
}

func (m M) DefaultGetFloat64(k string, f float64) float64 {
    v, ok := m.Get(k)
    if !ok {
        return f
    }
    return Float64(v)
}

func (m M) GetFloat64(k string) float64 {
    return m.DefaultGetFloat64(k, 0)
}

func (m M) GetFloat32(k string) float32 {
    v := m.GetFloat64(k)
    return float32(v)
}

// MVal 获取指定k的值
func (m M) MVal(k string) M {
    v, ok := m.Get(k)
    if !ok {
        return nil
    }
    nm := MVal(v)
    return nm
}

// SVal 获取指定k的Slice值
func (m M) SVal(k string) S {
    v, ok := m.Get(k)
    if !ok {
        return nil
    }
    s := SVal(v)
    return s
}

// UrlValues 转换为url.Values
func (m M) UrlValues() url.Values {
    uv := url.Values{}
    for k, v := range m {
        uv[k] = SVal(v).String()
    }
    return uv
}

// Map 获取原始数据
func (m M) Map() map[string]interface{} {
    return map[string]interface{}(m)
}

// Merge 合并数据
func (m M) Merge(n M) {
    for k, v := range n {
        if m.Contains(k) {
            continue
        }
        m[k] = v
    }
}

// ForceMerge 强制合并数据
func (m M) ForceMerge(n M) {
    for k, v := range n {
        m[k] = v
    }
}

// IsEmpty 判断是否为空
func (m M) IsEmpty(k string) bool {
    v, ok := m.Get(k)
    if !ok {
        return true
    }
    if v == nil {
        return true
    }
    return IsEmpty(v)
}

// UrlQuery 将map转换为query string
func (m M) UrlQuery() string {
    q := url.Values{}
    for k, v := range m {
        q.Set(k, String(v))
    }
    return q.Encode()
}

// Del 删除指定key
func (m M) Del(k string) {
    delete(m, k)
}

// ///////////////////// TYPE OF S ///////////////////////////

// S 常用的slice类型
type S []interface{}

// Load 从json中加载数据
func (s *S) Load(v []byte) error {
    if string(v) == "{}" {
        return nil
    }
    err := json.Unmarshal(v, &s)
    if err != nil {
        return err
    }
    return nil
}

// Json 将S转换为json字符串
func (s S) Json() string {
    v, _ := json.Marshal(s)
    return string(v)
}

func (s S) Size() int {
    return len(s)
}

// M 类型转换，将slice转换为map
func (s S) M() []M {
    size := s.Size()
    if size == 0 {
        return []M{}
    }
    v := make([]M, size)
    for i, s1 := range s {
        v[i] = MVal(s1)
    }
    return v
}

// Int64 类型转换，将所有元素转换成int64类型，并返回转换后的slice
func (s S) Int64() []int64 {
    size := s.Size()
    if size == 0 {
        return []int64{}
    }
    v := make([]int64, size)
    for i, s1 := range s {
        v[i] = Int64(s1)
    }
    return v
}

func (s S) Int() []int {
    size := s.Size()
    if size == 0 {
        return []int{}
    }
    v := make([]int, size)
    for i, s1 := range s {
        v[i] = Int(s1)
    }
    return v
}

func (s S) Uint() []uint {
    size := s.Size()
    if size == 0 {
        return []uint{}
    }
    v := make([]uint, size)
    for i, s1 := range s {
        v[i] = Uint(s1)
    }
    return v
}

func (s S) Uint64() []uint64 {
    size := s.Size()
    if size == 0 {
        return []uint64{}
    }
    v := make([]uint64, size)
    for i, s1 := range s {
        v[i] = Uint64(s1)
    }
    return v
}

func (s S) Float64() []float64 {
    size := s.Size()
    if size == 0 {
        return []float64{}
    }
    v := make([]float64, size)
    for i, s1 := range s {
        v[i] = Float64(s1)
    }
    return v
}

func (s S) String() []string {
    size := s.Size()
    if size == 0 {
        return []string{}
    }
    v := make([]string, size)
    for i, s1 := range s {
        v[i] = String(s1)
    }
    return v
}

// Filter 过滤，过滤掉slice中的空值对象
func (s S) Filter() {
    newS := S{}
    for _, v := range s {
        if IsEmpty(v) {
            continue
        }
        newS = append(newS, v)
    }
    s = newS
}

// Contains 包含检查
// 注意：1. 此方法因为在比较过程中进行了类型转换，因此存在风险；
//      2. 建议在明确数据内容的场景下使用此方法.
func (s S) Contains(v interface{}) bool {
    if len(s) == 0 {
        return false
    }
    switch v.(type) {
    case int, int64, int8, int16, int32:
        dstV := Int64(v)
        for _, sv := range s {
            if Int64(sv) == dstV {
                return true
            }
        }
    case uint, uint64, uint8, uint16, uint32, bool:
        dstV := Uint64(v)
        for _, sv := range s {
            if Uint64(sv) == dstV {
                return true
            }
        }
    case float32, float64:
        dstV := Float64(v)
        for _, sv := range s {
            if Float64(sv) == dstV {
                return true
            }
        }
    default:
        dstV := v.(string)
        for _, sv := range s {
            if String(sv) == dstV {
                return true
            }
        }
    }
    return false
}

// Interface 类型转换, to interface
func (s S) Interface() []interface{} {
    return []interface{}(s)
}

// Extend 对slice进行增强
func (s S) Extend(f func(interface{}) string) *XS {
    return XSVal(s, f)
}

// ///////////////////// TYPE OF XS - A EXTENDED SLICE ///////////////////////////

// XS 扩展增强的slice对象
type XS struct {
    sliceData []interface{}
    mapData   map[string]*xsDataWithCount
    keyFunc   func(interface{}) string // 定义将值转换为string的方法，主要用于去重
    mu        sync.RWMutex
}
type xsDataWithCount struct {
    data  interface{}
    count int
}

// NewXS 创建一个slice增强对象
func NewXS(f func(interface{}) string) *XS {
    xs := &XS{
        sliceData: make([]interface{}, 0),
        mapData:   make(map[string]*xsDataWithCount),
        keyFunc:   nil,
        mu:        sync.RWMutex{},
    }
    if f == nil {
        xs.keyFunc = func(i interface{}) string { // 定义一个默认方法
            return String(i)
        }
    } else {
        xs.keyFunc = f
    }
    return xs
}

// Size 获取元素列表个数
func (xs *XS) Size() int {
    return len(xs.sliceData)
}

// Add 添加（1个或多个）元素到列表
func (xs *XS) Add(args ...interface{}) {
    if len(args) == 0 {
        return
    }
    xs.mu.Lock()
    defer xs.mu.Unlock()
    xs.sliceData = append(xs.sliceData, args...)
    for _, v := range args {
        k := xs.keyFunc(v)
        if _, ok := xs.mapData[k]; !ok {
            xs.mapData[k] = &xsDataWithCount{
                data:  v,
                count: 0,
            }
        }
        xs.mapData[k].count++
    }
}

// SAdd 添加（1个或多个）元素到列表, 如果元素已经存在，则忽略该元素
func (xs *XS) SAdd(args ...interface{}) {
    if len(args) == 0 {
        return
    }
    xs.mu.Lock()
    defer xs.mu.Unlock()
    for _, v := range args {
        k := xs.keyFunc(v)
        if _, ok := xs.mapData[k]; !ok {
            xs.sliceData = append(xs.sliceData, v)
            xs.mapData[k] = &xsDataWithCount{
                data:  v,
                count: 1,
            }
        }
    }
}

// Delete 删除指定值的元素
func (xs *XS) Delete(arg interface{}) {
    if !xs.Contains(arg) {
        return
    }
    key := xs.keyFunc(arg)
    xs.mu.Lock()
    defer xs.mu.Unlock()
    size := len(xs.sliceData)
    for i := 0; i < size; i++ {
        v := xs.sliceData[i]
        k := xs.keyFunc(v)
        if k == key {
            // 删除元素
            for j := i + 1; j < size; j++ {
                xs.sliceData[j-1] = xs.sliceData[j]
            }
            size--
            xs.mapData[k].count--
            if xs.mapData[k].count == 0 {
                // 元素已删除完毕
                delete(xs.mapData, key)
                break
            }
        }
    }
}

// Contains 检查是否包含某个元素
func (xs *XS) Contains(v interface{}) bool {
    k := xs.keyFunc(v)
    xs.mu.RLock()
    defer xs.mu.RUnlock()
    if dc, ok := xs.mapData[k]; ok && dc.count > 0 {
        return true
    }
    return false
}

// Unique 对列表元素去重
func (xs *XS) Unique() []interface{} {
    s := make([]interface{}, len(xs.mapData))
    xs.mu.RLock()
    defer xs.mu.RUnlock()
    i := 0
    for _, v := range xs.mapData {
        s[i] = v.data
        i++
    }
    return s
}

// UniqueFilter 对列表元素去重，并根据filter方法进行过滤
// filter参数是一个字符串，此值是根据创建XS对象时指定的keyFunc得到的
func (xs *XS) UniqueFilter(f func(string) bool) []interface{} {
    s := make([]interface{}, 0)
    xs.mu.RLock()
    defer xs.mu.RUnlock()
    for k, v := range xs.mapData {
        if ok := f(k); ok {
            s = append(s, v.data)
        }
    }
    return s
}

// Slice 获取切片(为了防止修改对象本身，需单独复制一个切片出来)
func (xs *XS) Slice() []interface{} {
    s := make([]interface{}, xs.Size())
    copy(s, xs.sliceData)
    return s
}

// Intersect 计算给定切片与当前列表的交集
func (xs *XS) Intersect(arrs ...S) *XS {
    // 创建一个新的XS对象
    newXs := NewXS(xs.keyFunc)
    newXs.Add(xs.Slice()...)
    // 计算目标匹配次数
    checkCount := len(arrs) + 1
    for _, arr := range arrs {
        newXs.Add(arr.Interface()...)
    }
    // 提取结果
    outS := S{}
    for _, dc := range newXs.mapData {
        if dc.count == checkCount {
            outS = append(outS, dc.data)
        }
    }
    // 返回结果
    return outS.Extend(xs.keyFunc)
}

// Union 计算给定切片与当前列表的并集
func (xs *XS) Union(arrs ...S) *XS {
    // 创建一个新的xs对象,，并以原来的xs数据填充
    newXs := NewXS(xs.keyFunc)
    newXs.Add(xs.Slice()...)
    // 添加求并集的数据
    for _, arr := range arrs {
        newXs.Add(arr.Interface()...)
    }
    // 提取结果并返回
    return S(newXs.Unique()).Extend(xs.keyFunc)
}

// ///////////////////// TYPE OF TreeItem ///////////////////////////

// TreeItem 定义构成树的元素接口
type TreeItem interface {
    Key() string
    ParentKey() string
    AddChild(ti TreeItem)
    GetChild(k string) (TreeItem, bool)
}

// Tree 定义一个树结构
type Tree struct {
    items   map[string]TreeItem
    keyMaps map[string]string
    keys    []string // 用于记录输出顺序
}

// NewTree 创建一个树实例
func NewTree() *Tree {
    return &Tree{
        items:   make(map[string]TreeItem, 0),
        keyMaps: make(map[string]string, 0),
        keys:    make([]string, 0),
    }
}

// path 返回一个元素的层级路径关系
func (tree *Tree) path(pk string) []string {
    paths := []string{pk}
    for {
        _pk, ok := tree.keyMaps[pk]
        if !ok {
            break
        }
        paths = append(paths, _pk)
        pk = _pk
    }
    return paths
}

// Add 添加Tree Items
func (tree *Tree) Add(items ...TreeItem) {
    if len(items) == 0 {
        return
    }
    // 第一次循环，将所有数据加入map
    for _, item := range items {
        k := item.Key()
        pk := item.ParentKey()
        tree.items[k] = item
        if pk == "" {
            tree.keys = append(tree.keys, k)
        } else {
            tree.keyMaps[k] = pk
        }
    }
}

// build 构造树
func (tree *Tree) build() {
    for k, item := range tree.items {
        pk := item.ParentKey()
        // pk 为空表示为顶级
        if pk == "" {
            continue
        }
        // 按层级处理, 获取元素的逐级上级key
        // 即: 从最里层到最外层级，但是在添加元素时应该反向操作
        paths := tree.path(pk)
        size := len(paths)
        root, ok := tree.items[paths[size-1]]
        if !ok {
            // root不存在，删除当前元素
            delete(tree.items, k)
            break
        }
        if size < 2 {
            root.AddChild(item)
        } else {
            parent := root
            for i := size - 2; i >= 0; i-- {
                current, ok := parent.GetChild(paths[i])
                if !ok {
                    // 下级元素不存在，查看数据池（map）中是否存在该元素
                    if child, ok := tree.items[paths[i]]; ok {
                        parent.AddChild(child)
                        current = child
                    }
                }
                if current == nil {
                    // 下级元素不存在也没有找到
                    // 意味着当前元素的关系是不完整的（因为可以查找到路径，所以理论上不应该存在这种情况）
                    delete(tree.items, k)
                    break
                }
                parent = current
                // 找到直接上级
                if i == 0 {
                    // 添加当前元素
                    current.AddChild(item)
                    delete(tree.items, k)
                    continue
                }
            }
        }
    }
}

// Tree 返回树
func (tree *Tree) Tree() []TreeItem {
    tree.build()
    items := make([]TreeItem, 0)
    for _, k := range tree.keys {
        if item, ok := tree.items[k]; ok {
            items = append(items, item)
        }
    }
    return items
}
