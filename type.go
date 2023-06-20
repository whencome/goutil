package gotil

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/url"
    "reflect"
    "sort"
    "strings"
)

/////////////////////// COMMON FUNCS ////////////////////////

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

//////////////////// TYPE OF KVPairs  ///////////////////////
// KVPair 参数键值对
type KVPair struct {
    K, V string
}

// KVPairs 参数键值对列表
type KVPairs []KVPair

// 上西安排序接口
func (p KVPairs) Less(i, j int) bool {
    return p[i].K < p[j].K
}

func (p KVPairs) Swap(i, j int) {
    p[i], p[j] = p[j], p[i]
}

func (p KVPairs) Len() int {
    return len(p)
}

// 排序
func (p KVPairs) Sort() {
    sort.Sort(p)
}

// 倒序
func (p KVPairs) Reverse() {
    sort.Reverse(p)
}

// 添加键值对
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

/////////////////////// TYPE OF M ///////////////////////////

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

// UrlQuery 判断是否为空
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

/////////////////////// TYPE OF S ///////////////////////////

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

// Json 获取json格式数据
func (s S) Json() string {
    v, _ := json.Marshal(s)
    return string(v)
}

func (s S) Size() int {
    return len(s)
}

// M 类型转换
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

// Int64 类型转换
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
