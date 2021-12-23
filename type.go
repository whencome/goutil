package gotil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

/////////////////////// COMMON FUNCS ////////////////////////
// GetM 将给定V转换成一个M对象，如果不支持转换，则直接返回空的M值
func GetM(v interface{}) M {
	var nm M = M{}
	switch v.(type) {
	case map[string]interface{}:
		mv, _ := v.(map[string]interface{})
		if mv != nil {
			nm = M(mv)
		}
	case map[string]string:
		mv, _ := v.(map[string]string)
		if len(mv) > 0 {
			for f, d := range mv {
				nm[f] = d
			}
		}
	case map[string]int:
		mv, _ := v.(map[string]int)
		if len(mv) > 0 {
			for f, d := range mv {
				nm[f] = d
			}
		}
	case map[string]int64:
		mv, _ := v.(map[string]int64)
		if len(mv) > 0 {
			for f, d := range mv {
				nm[f] = d
			}
		}
	case map[string]uint:
		mv, _ := v.(map[string]uint)
		if len(mv) > 0 {
			for f, d := range mv {
				nm[f] = d
			}
		}
	case map[string]uint64:
		mv, _ := v.(map[string]uint64)
		if len(mv) > 0 {
			for f, d := range mv {
				nm[f] = d
			}
		}
	case map[string]float32:
		mv, _ := v.(map[string]float32)
		if len(mv) > 0 {
			for f, d := range mv {
				nm[f] = d
			}
		}
	case map[string]float64:
		mv, _ := v.(map[string]float64)
		if len(mv) > 0 {
			for f, d := range mv {
				nm[f] = d
			}
		}
	case map[string]bool:
		mv, _ := v.(map[string]bool)
		if len(mv) > 0 {
			for f, d := range mv {
				nm[f] = d
			}
		}
	case string:
		_ = json.Unmarshal([]byte(v.(string)), &nm)
	}
	return nm
}

// GetS 将v转换为一个Slice值，如果不支持转换，则返回空，使用者需要预先确认值类型
func GetS(v interface{}) S {
	s := S{}
	switch v.(type) {
	case []interface{}:
		s = S(v.([]interface{}))
	case []string:
		sv, _ := v.([]string)
		if len(sv) > 0 {
			for _, d := range sv {
				s = append(s, d)
			}
		}
	case []int:
		iv, _ := v.([]int)
		if len(iv) > 0 {
			for _, d := range iv {
				s = append(s, d)
			}
		}
	case []int64:
		iv, _ := v.([]int64)
		if len(iv) > 0 {
			for _, d := range iv {
				s = append(s, d)
			}
		}
	case []uint:
		iv, _ := v.([]uint)
		if len(iv) > 0 {
			for _, d := range iv {
				s = append(s, d)
			}
		}
	case []uint64:
		iv, _ := v.([]uint64)
		if len(iv) > 0 {
			for _, d := range iv {
				s = append(s, d)
			}
		}
	case []float32:
		fv, _ := v.([]float32)
		if len(fv) > 0 {
			for _, d := range fv {
				s = append(s, d)
			}
		}
	case []float64:
		fv, _ := v.([]float64)
		if len(fv) > 0 {
			for _, d := range fv {
				s = append(s, d)
			}
		}
	default:
		s = S([]interface{}{v})
	}
	return s
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
	sKeys := GetS(keys)
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

// 将键值对转换成xml
func (p KVPairs) ToXML() string {
	buf := bytes.Buffer{}
	buf.WriteString("<xml>")
	for _, pair := range p {
		buf.WriteString(fmt.Sprintf("<%s><![CDATA[%s]]></%s>", pair.K, pair.V, pair.K))
	}
	buf.WriteString("</xml>")
	return buf.String()
}

// 将键值对转换为Map
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
// 常用的map类型
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

// GetBool 获取配置项k的值，返回字符串
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

// GetM 获取指定k的值
func (m M) GetM(k string) M {
	v, ok := m.Get(k)
	if !ok {
		return nil
	}
	nm := GetM(v)
	return nm
}

// GetS 获取指定k的Slice值
func (m M) GetS(k string) S {
	v, ok := m.Get(k)
	if !ok {
		return nil
	}
	s := GetS(v)
	return s
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

// 强制合并数据
func (m M) ForceMerge(n M) {
	for k, v := range n {
		m[k] = v
	}
}

// 判断是否为空
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

// 判断是否为空
func (m M) UrlQuery() string {
	q := url.Values{}
	for k, v := range m {
		q.Set(k, String(v))
	}
	return q.Encode()
}

// 删除指定key
func (m M) Del(k string) {
	delete(m, k)
}

/////////////////////// TYPE OF S ///////////////////////////
// 常用的slice类型
type S []interface{}

// 从json中加载数据
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

// 类型转换
func (s S) M() []M {
	size := s.Size()
	if size == 0 {
		return []M{}
	}
	v := make([]M, size)
	for i, s1 := range s {
		v[i] = GetM(s1)
	}
	return v
}

// 类型转换
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

// 过滤
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

// 包含检查
// 注意：此方法可能存在一定风险
func (s S) Contains(v interface{}) bool {
	if len(s) == 0 {
		return false
	}
	switch v.(type) {
	case int, int64, int8, int16, int32:
		dstV := Int64(v)
		srcS := s.Int64()
		for _, sv := range srcS {
			if sv == dstV {
				return true
			}
		}
	case uint, uint64, uint8, uint16, uint32, bool:
		dstV := Uint64(v)
		srcS := s.Uint64()
		for _, sv := range srcS {
			if sv == dstV {
				return true
			}
		}
	case float32, float64:
		dstV := Float64(v)
		srcS := s.Float64()
		for _, sv := range srcS {
			if sv == dstV {
				return true
			}
		}
	default:
		dstV := v.(string)
		srcS := s.String()
		for _, sv := range srcS {
			if sv == dstV {
				return true
			}
		}
	}
	return false
}

/////////////////////// TYPE OF O（Object） ///////////////////////////
// Value 定义一个通用的Value结构体，用于统一处理类型转换
type obj struct {
	v interface{}
}
type O obj

func Object(v interface{}) O {
	return O{
		v: v,
	}
}

// IsNil 判断值是否为nil
func (o O) IsNil() bool {
	return IsNil(o.v)
}

// IsEmpty 判断值是否为空
func (o O) IsEmpty() bool {
	return IsEmpty(o.v)
}

// String Get a string value of o
func (o O) String() string {
	var strVal = ""
	switch o.v.(type) {
	case int, int8, int16, int32, int64:
		n := o.Int64()
		strVal = strconv.FormatInt(n, 10)
	case uint, uint8, uint16, uint32, uint64:
		n := o.Uint64()
		strVal = strconv.FormatUint(n, 10)
	case float32:
		strVal = strconv.FormatFloat(float64(o.v.(float32)), 'f', -1, 64)
	case float64:
		strVal = strconv.FormatFloat(o.v.(float64), 'f', -1, 64)
	case string:
		strVal = o.v.(string)
	case []byte:
		strVal = string(o.v.([]byte))
	case []rune:
		strVal = string(o.v.([]rune))
	case bool:
		strVal = strconv.FormatBool(o.v.(bool))
	default:
		if o.IsNil() {
			strVal = ""
		} else {
			strVal = fmt.Sprint(o.v)
		}
	}
	return strVal
}

// SqlValue 获取插入数据库需要的值
func (o O) SqlValue() string {
	var strVal = ""
	switch o.v.(type) {
	case int, int8, int16, int32, int64:
		n := o.Int64()
		strVal = strconv.FormatInt(n, 10)
	case uint, uint8, uint16, uint32, uint64:
		n := o.Uint64()
		strVal = strconv.FormatUint(n, 10)
	case float32:
		strVal = strconv.FormatFloat(float64(o.v.(float32)), 'f', -1, 64)
	case float64:
		strVal = strconv.FormatFloat(o.v.(float64), 'f', -1, 64)
	case string:
		strVal = o.v.(string)
	case []byte:
		strVal = string(o.v.([]byte))
	case []rune:
		strVal = string(o.v.([]rune))
	case bool:
		strVal = "0"
		if o.v.(bool) {
			strVal = "1"
		}
	default:
		strVal = fmt.Sprint(o.v)
	}
	// 对特殊字符进行处理
	strVal = EscapeSqlValue(strVal)
	// 返回结果
	return strVal
}

// Int64 get int64 value
func (o O) Int64() int64 {
	switch o.v.(type) {
	case int:
		return int64(o.v.(int))
	case int8:
		return int64(o.v.(int8))
	case int16:
		return int64(o.v.(int16))
	case int32:
		return int64(o.v.(int32))
	case int64:
		return o.v.(int64)
	case uint:
		return int64(o.v.(uint))
	case uint8:
		return int64(o.v.(uint8))
	case uint16:
		return int64(o.v.(uint16))
	case uint32:
		return int64(o.v.(uint32))
	case uint64:
		return int64(o.v.(uint64))
	case float32:
		return int64(o.v.(float32))
	case float64:
		return int64(o.v.(float64))
	case string:
		n, err := strconv.ParseInt(string(o.v.(string)), 10, 64)
		if err != nil {
			return 0
		}
		return n
	case []byte:
		n, err := strconv.ParseInt(string(o.v.([]byte)), 10, 64)
		if err != nil {
			return 0
		}
		return n
	case []rune:
		n, err := strconv.ParseInt(string(o.v.([]rune)), 10, 64)
		if err != nil {
			return 0
		}
		return n
	case bool:
		intVal := int64(0)
		if o.v.(bool) {
			intVal = 1
		}
		return intVal
	default:
		return 0
	}
	return 0
}

// Uint64 get uint64 value
func (o O) Uint64() uint64 {
	switch o.v.(type) {
	case int:
		return uint64(o.v.(int))
	case int8:
		return uint64(o.v.(int8))
	case int16:
		return uint64(o.v.(int16))
	case int32:
		return uint64(o.v.(int32))
	case int64:
		return uint64(o.v.(int64))
	case uint:
		return uint64(o.v.(uint))
	case uint8:
		return uint64(o.v.(uint8))
	case uint16:
		return uint64(o.v.(uint16))
	case uint32:
		return uint64(o.v.(uint32))
	case uint64:
		return o.v.(uint64)
	case float32:
		return uint64(o.v.(float32))
	case float64:
		return uint64(o.v.(float64))
	case string:
		n, err := strconv.ParseUint(string(o.v.(string)), 10, 64)
		if err != nil {
			return 0
		}
		return n
	case []byte:
		n, err := strconv.ParseUint(string(o.v.([]byte)), 10, 64)
		if err != nil {
			return 0
		}
		return n
	case []rune:
		n, err := strconv.ParseUint(string(o.v.([]rune)), 10, 64)
		if err != nil {
			return 0
		}
		return n
	case bool:
		intVal := uint64(0)
		if o.v.(bool) {
			intVal = 1
		}
		return intVal
	default:
		return 0
	}
	return 0
}

// Float64 get float64 value
func (o O) Float64() float64 {
	switch o.v.(type) {
	case int:
		return float64(o.v.(int))
	case int8:
		return float64(o.v.(int8))
	case int16:
		return float64(o.v.(int16))
	case int32:
		return float64(o.v.(int32))
	case int64:
		return float64(o.v.(int64))
	case uint:
		return float64(o.v.(uint))
	case uint8:
		return float64(o.v.(uint8))
	case uint16:
		return float64(o.v.(uint16))
	case uint32:
		return float64(o.v.(uint32))
	case uint64:
		return float64(o.v.(uint64))
	case float32:
		return float64(o.v.(float32))
	case float64:
		return float64(o.v.(float64))
	case string:
		n, err := strconv.ParseFloat(string(o.v.(string)), 64)
		if err != nil {
			return 0
		}
		return n
	case []byte:
		n, err := strconv.ParseFloat(string(o.v.([]byte)), 64)
		if err != nil {
			return 0
		}
		return n
	case []rune:
		n, err := strconv.ParseFloat(string(o.v.([]rune)), 64)
		if err != nil {
			return 0
		}
		return n
	case bool:
		n := float64(0)
		if o.v.(bool) {
			n = 1
		}
		return n
	default:
		return 0
	}
	return 0
}

// Boolean get bool value
func (o O) Boolean() bool {
	v := o.Uint64()
	if v > 0 {
		return true
	}
	return false
}
