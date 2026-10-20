package goutil

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
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() == reflect.Map {
		for _, k := range value.MapKeys() {
			sk := String(k.Interface())
			mv[sk] = value.MapIndex(k).Interface()
		}
	} /* else if value.Kind() == reflect.String {
		// 使用此功能需自行承担失败风险
		_ = json.Unmarshal([]byte(v.(string)), &mv)
	}*/
	return mv
}

// SVal 将v转换为一个Slice值，如果不支持转换，则返回空，使用者需要预先确认值类型
func SVal(v interface{}) S {
	sv := S{}
	if v == nil {
		return sv
	}
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
		for i := 0; i < value.Len(); i++ {
			sv = append(sv, value.Index(i).Interface())
		}
	}
	return sv
}

// SMVal 将v转换为一个M列表
func SMVal(v interface{}) []M {
	smv := make([]M, 0)
	if v == nil {
		return smv
	}
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
		for i := 0; i < value.Len(); i++ {
			smv = append(smv, MVal(value.Index(i).Interface()))
		}
	}
	return smv
}

// XSVal 将v转换为一个增强的Slice对象，如果不支持转换，则返回空，使用者需要预先确认值类型
func XSVal(v interface{}, f func(interface{}) string) *XS {
	if xs, ok := v.(*XS); ok {
		return xs
	}
	xs := NewXS(f)
	if v == nil {
		return xs
	}
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			xs.Add(value.Index(i).Interface())
		}
	}
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
// 说明：历史实现未真正执行排序（sort.Reverse的返回值被丢弃），此处已修正为就地倒序
func (p KVPairs) Reverse() {
	sort.Sort(sort.Reverse(p))
}

// Add 添加键值对
// 注意：由于KVPairs是切片类型，且此方法使用值接收者，append的结果不会回写到调用方，
// 因此此方法实际上不会修改原列表（历史遗留缺陷，为保证兼容性保留）。
// 请使用 Append 方法，或直接使用 append(pairs, KVPair{K: k, V: v})。
// Deprecated: 此方法无实际效果，请使用 Append。
func (p KVPairs) Add(k, v string) {
	p = append(p, KVPair{K: k, V: v})
}

// Append 添加一个键值对，并返回新的键值对列表（不会修改原列表）
func (p KVPairs) Append(k, v string) KVPairs {
	newPairs := make(KVPairs, 0, len(p)+1)
	newPairs = append(newPairs, p...)
	return append(newPairs, KVPair{K: k, V: v})
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

// Keys 获取map的键列表
func (m M) Keys() []string {
	keys := make([]string, len(m))
	if len(m) == 0 {
		return keys
	}
	i := 0
	for k, _ := range m {
		keys[i] = k
		i++
	}
	return keys
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

// SMVal 获取一个M切片
func (m M) SMVal(k string) []M {
	v, ok := m.Get(k)
	if !ok {
		return nil
	}
	s := SVal(v)
	if len(s) == 0 {
		return nil
	}
	sm := make([]M, len(s))
	for i, v := range s {
		sm[i] = MVal(v)
	}
	return sm
}

// UrlValues 转换为url.Values
// 说明：数据类型为切片时按多值处理，其它类型按单值处理；
// 历史实现对于非切片类型的值会返回空字符串，此处已修正
func (m M) UrlValues() url.Values {
	uv := url.Values{}
	for k, v := range m {
		if IsList(v) {
			uv[k] = SVal(v).String()
			continue
		}
		uv[k] = []string{String(v)}
	}
	return uv
}

// Map 获取原始数据
func (m M) Map() map[string]interface{} {
	return map[string]interface{}(m)
}

// Merge 合并数据（深拷贝）
func (m M) Merge(n M) {
	for k, v := range n {
		if m.Contains(k) {
			continue
		}
		m[k] = DeepCopy(v)
	}
}

// ForceMerge 强制合并数据（深拷贝）
func (m M) ForceMerge(n M) {
	for k, v := range n {
		m[k] = DeepCopy(v)
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

// IsArray 判断指定的值是否是切片或者数组
func (m M) IsArray(k string) bool {
	v, ok := m.Get(k)
	if !ok {
		return false
	}
	value := reflect.ValueOf(v)
	return value.Kind() == reflect.Slice || value.Kind() == reflect.Array
}

// UrlQuery 将map转换为query string
// 说明：数据类型为切片时会转换为重复的key（如 a=1&a=2），其它类型按单值处理
func (m M) UrlQuery() string {
	q := url.Values{}
	for k, v := range m {
		if IsList(v) {
			for _, item := range SVal(v) {
				q.Add(k, String(item))
			}
			continue
		}
		q.Set(k, String(v))
	}
	return q.Encode()
}

// Del 删除指定key
func (m M) Del(k string) {
	delete(m, k)
}

// GetByPath 根据path获取值
// path使用"."分隔，如 "a.b.c"；路径中的每一级都必须是一个map，否则返回false
// 说明：此方法为只读操作，不会复制数据；历史实现每次调用都会深拷贝一份数据，性能较差
func (m M) GetByPath(path string) (interface{}, bool) {
	if path == "" {
		return nil, false
	}
	var current interface{} = m
	for _, part := range strings.Split(path, ".") {
		v, ok := mapIndex(current, part)
		if !ok {
			return nil, false
		}
		current = v
	}
	return current, true
}

// mapIndex 从map中获取指定key的值
// 说明：支持任意value类型的map，也支持非string类型的key（会先转换成string再比较）
func mapIndex(v interface{}, key string) (interface{}, bool) {
	val := reflect.ValueOf(v)
	for val.IsValid() && (val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface) {
		if val.IsNil() {
			return nil, false
		}
		val = val.Elem()
	}
	if !val.IsValid() || val.Kind() != reflect.Map {
		return nil, false
	}
	keyType := val.Type().Key()
	if keyType.Kind() == reflect.String {
		mv := val.MapIndex(reflect.ValueOf(key).Convert(keyType))
		if !mv.IsValid() {
			return nil, false
		}
		return mv.Interface(), true
	}
	// 非string类型的key，退化为遍历比较
	for _, k := range val.MapKeys() {
		if String(k.Interface()) == key {
			return val.MapIndex(k).Interface(), true
		}
	}
	return nil, false
}

// MValByPath 根据路径获取一个M值
func (m M) MValByPath(path string) (M, bool) {
	v, ok := m.GetByPath(path)
	if !ok || !IsMap(v) {
		return nil, false
	}
	return MVal(v), true
}

// SValByPath 根据路径获取一个S值
func (m M) SValByPath(path string) (S, bool) {
	v, ok := m.GetByPath(path)
	if !ok || !IsList(v) {
		return nil, false
	}
	return SVal(v), true
}

// SMValByPath 根据路径获取一个SM值
func (m M) SMValByPath(path string) ([]M, bool) {
	v, ok := m.GetByPath(path)
	if !ok || !IsList(v) {
		return nil, false
	}
	sm := make([]M, 0)
	list := SVal(v).Interface()
	for _, val := range list {
		if !IsMap(val) {
			return nil, false
		}
		sm = append(sm, MVal(val))
	}
	return sm, true
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
// 注意：由于S是切片类型，且此方法使用值接收者，方法内部对s的重新赋值不会影响调用方，
// 因此此方法实际上不会修改原列表（历史遗留缺陷，为保证兼容性保留）。
// 请使用 Filtered 方法获取过滤后的新列表。
// Deprecated: 此方法无实际效果，请使用 Filtered。
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

// Filtered 过滤，返回一个去除了空值对象的新列表，原列表不受影响
func (s S) Filtered() S {
	newS := make(S, 0, len(s))
	for _, v := range s {
		if IsEmpty(v) {
			continue
		}
		newS = append(newS, v)
	}
	return newS
}

// Contains 包含检查
// 注意：1. 此方法因为在比较过程中进行了类型转换，因此存在风险；
//  2. 建议在明确数据内容的场景下使用此方法.
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
		// 注意：此处不能直接使用 v.(string) 断言，否则传入非字符串类型时会panic
		dstV := String(v)
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
	xs.mu.RLock()
	defer xs.mu.RUnlock()
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
// 说明：删除列表中所有key与被删除值相同的元素（key由创建XS时指定的keyFunc生成）；
// 若元素不存在则不做任何操作。此方法为并发安全方法。
func (xs *XS) Delete(arg interface{}) {
	key := xs.keyFunc(arg)
	xs.mu.Lock()
	defer xs.mu.Unlock()
	if dc, ok := xs.mapData[key]; !ok || dc.count <= 0 {
		return
	}
	// 重新构建切片，避免遗留脏数据
	newSlice := make([]interface{}, 0, len(xs.sliceData))
	for _, v := range xs.sliceData {
		if xs.keyFunc(v) == key {
			continue
		}
		newSlice = append(newSlice, v)
	}
	xs.sliceData = newSlice
	delete(xs.mapData, key)
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
// 说明：返回结果按照元素首次加入的顺序排列，保证结果的稳定性
func (xs *XS) Unique() []interface{} {
	xs.mu.RLock()
	defer xs.mu.RUnlock()
	s := make([]interface{}, 0, len(xs.mapData))
	added := make(map[string]struct{}, len(xs.mapData))
	for _, v := range xs.sliceData {
		k := xs.keyFunc(v)
		if _, ok := added[k]; ok {
			continue
		}
		added[k] = struct{}{}
		s = append(s, v)
	}
	return s
}

// UniqueFilter 对列表元素去重，并根据filter方法进行过滤
// filter参数是一个字符串，此值是根据创建XS对象时指定的keyFunc得到的
// 说明：返回结果按照元素首次加入的顺序排列，保证结果的稳定性
func (xs *XS) UniqueFilter(f func(string) bool) []interface{} {
	xs.mu.RLock()
	defer xs.mu.RUnlock()
	s := make([]interface{}, 0, len(xs.mapData))
	added := make(map[string]struct{}, len(xs.mapData))
	for _, v := range xs.sliceData {
		k := xs.keyFunc(v)
		if _, ok := added[k]; ok {
			continue
		}
		if f == nil || f(k) {
			added[k] = struct{}{}
			s = append(s, v)
		}
	}
	return s
}

// Slice 获取切片(为了防止修改对象本身，需单独复制一个切片出来)
func (xs *XS) Slice() []interface{} {
	xs.mu.RLock()
	defer xs.mu.RUnlock()
	s := make([]interface{}, len(xs.sliceData))
	copy(s, xs.sliceData)
	return s
}

// Intersect 计算给定切片与当前列表的交集
// 说明：返回同时存在于当前列表与所有给定切片中的元素（按key去重，保留当前列表中的顺序）。
// 若未传入任何切片，则返回当前列表去重后的结果。
func (xs *XS) Intersect(arrs ...S) *XS {
	outS := S{}
	if len(arrs) == 0 {
		return S(xs.Unique()).Extend(xs.keyFunc)
	}
	// 统计当前列表中每个key首次出现的值
	order := make([]string, 0)
	dataByKey := make(map[string]interface{})
	for _, v := range xs.Unique() {
		k := xs.keyFunc(v)
		if _, ok := dataByKey[k]; ok {
			continue
		}
		dataByKey[k] = v
		order = append(order, k)
	}
	// 统计每个key在多少个集合中出现（每个集合内部先去重）
	present := make(map[string]int, len(dataByKey))
	for k := range dataByKey {
		present[k] = 1
	}
	for _, arr := range arrs {
		seen := make(map[string]struct{}, len(arr))
		for _, v := range arr {
			k := xs.keyFunc(v)
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			if _, ok := present[k]; ok {
				present[k]++
			}
		}
	}
	// 提取结果
	target := len(arrs) + 1
	for _, k := range order {
		if present[k] == target {
			outS = append(outS, dataByKey[k])
		}
	}
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
	items    map[string]TreeItem
	keyMaps  map[string]string
	keys     []string // 用于记录顶级节点的输出顺序
	order    []string // 用于记录所有节点的加入顺序，保证构建结果的顺序稳定
	rootKeys map[string]struct{}
}

// NewTree 创建一个树实例
func NewTree() *Tree {
	return &Tree{
		items:    make(map[string]TreeItem, 0),
		keyMaps:  make(map[string]string, 0),
		keys:     make([]string, 0),
		order:    make([]string, 0),
		rootKeys: make(map[string]struct{}, 0),
	}
}

// path 返回一个元素的层级路径关系
// 说明：返回顺序为[父级, 父级的父级, ..., 根]，若父子关系中存在环，则在遇到重复key时终止，不会死循环
func (tree *Tree) path(pk string) []string {
	paths := []string{pk}
	visited := map[string]struct{}{pk: {}}
	for {
		_pk, ok := tree.keyMaps[pk]
		if !ok {
			break
		}
		if _, exists := visited[_pk]; exists {
			// 父级关系存在环，终止查找
			break
		}
		visited[_pk] = struct{}{}
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
	if tree.items == nil {
		tree.items = make(map[string]TreeItem)
	}
	if tree.keyMaps == nil {
		tree.keyMaps = make(map[string]string)
	}
	if tree.rootKeys == nil {
		tree.rootKeys = make(map[string]struct{})
	}
	// 第一次循环，将所有数据加入map
	for _, item := range items {
		if item == nil {
			continue
		}
		k := item.Key()
		pk := item.ParentKey()
		if _, exists := tree.items[k]; !exists {
			// 记录加入顺序，保证输出结果的顺序稳定
			tree.order = append(tree.order, k)
		}
		tree.items[k] = item
		if pk == "" {
			// 避免同一个顶级节点被重复加入
			if _, exists := tree.rootKeys[k]; !exists {
				tree.rootKeys[k] = struct{}{}
				tree.keys = append(tree.keys, k)
			}
		} else {
			tree.keyMaps[k] = pk
			if _, exists := tree.rootKeys[k]; exists {
				// 由顶级节点变为子节点，需要从顶级列表中移除
				delete(tree.rootKeys, k)
				tree.removeRootKey(k)
			}
		}
	}
}

// removeRootKey 从顶级key列表中移除指定的key
func (tree *Tree) removeRootKey(k string) {
	for i, key := range tree.keys {
		if key == k {
			tree.keys = append(tree.keys[:i], tree.keys[i+1:]...)
			return
		}
	}
}

// build 构造树
// 说明：按元素加入的顺序处理，保证子节点的顺序稳定；父级不存在的元素会被忽略（从结果中移除）
func (tree *Tree) build() {
	// 记录已经挂载到树上的节点，避免同一个节点被重复挂载（例如先作为中间层级挂载，之后又被单独处理）
	attached := make(map[string]struct{}, len(tree.items))
	for _, k := range tree.order {
		item, exists := tree.items[k]
		if !exists {
			continue
		}
		if _, ok := attached[k]; ok {
			continue
		}
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
			continue
		}
		if size == 1 {
			// 直接挂到顶级节点下
			root.AddChild(item)
			delete(tree.items, k)
			continue
		}
		// 从顶级节点逐级向下查找，补齐中间层级后挂载当前节点
		parent := root
		complete := true
		for i := size - 2; i >= 0; i-- {
			childKey := paths[i]
			current, found := parent.GetChild(childKey)
			if !found {
				// 下级元素不存在，查看数据池（map）中是否存在该元素
				child, inPool := tree.items[childKey]
				if !inPool {
					// 关系不完整，丢弃当前元素
					complete = false
					break
				}
				parent.AddChild(child)
				// 标记为已挂载，并从数据池中移除，避免后续重复挂载；
				// 后续节点在查找该中间节点时，会通过父节点的GetChild找到它
				attached[childKey] = struct{}{}
				delete(tree.items, childKey)
				current = child
			}
			if current == nil {
				complete = false
				break
			}
			parent = current
		}
		if !complete {
			delete(tree.items, k)
			continue
		}
		parent.AddChild(item)
		delete(tree.items, k)
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
