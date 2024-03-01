package sqlcond

import (
    "bytes"
    "errors"
    "fmt"
    "regexp"
    "strings"

    "github.com/whencome/goutil"
)

const (
    LogicAnd       = "AND"
    LogicOr        = "OR"
    keywordPattern = `，|；|。|\.|\#|\-|\,|\;|\_|\'|\"|\%|\(|\)|\{|\}|\[|\]`
)

var (
    ErrInvalidCondition  = errors.New("invalid query condition")
    ErrNotSupportedQuery = errors.New("query logic not supported")
    // ErrUselessQuery 无用查询，即根据条件或构造条件时知道查询结果为空，此情况下没有查询结果
    ErrUselessQuery = errors.New("useless query")
)

// Condition 定义一个查询条件
type Condition struct {
    logic string
    conds []interface{}
    Error error // 记录第一个错误信息
}

// subcond 定义子条件
type subcond struct {
    command string
    values  []interface{}
}

// Option 用于查询条件参数设置
type Option func(*Condition)

// WithAndLogic 返回一个设置逻辑为AND的Option
func WithAndLogic() Option {
    return func(c *Condition) {
        c.logic = LogicAnd
    }
}

// WithOrLogic 返回一个设置逻辑为OR的Option
func WithOrLogic() Option {
    return func(c *Condition) {
        c.logic = LogicOr
    }
}

// WithLogic customize logic
func WithLogic(logic string) Option {
    return func(c *Condition) {
        if logic != LogicAnd && logic != LogicOr {
            logic = LogicAnd
        }
        c.logic = logic
    }
}

// New 创建一个查询条件
func New(opts ...Option) *Condition {
    c := &Condition{
        logic: LogicAnd,
        conds: make([]interface{}, 0),
        Error: nil,
    }
    for _, opt := range opts {
        opt(c)
    }
    return c
}

// WithClause 直接根据条件构造一个Condition，相当于New与Add的合并
func WithClause(opts ...interface{}) *Condition {
    cond := New()
    cond.Add(opts...)
    return cond
}

// Add 添加一个或者多个查询条件
func (c *Condition) Add(opts ...interface{}) {
    if len(opts) == 0 || c.Error != nil {
        return
    }
    cmd := opts[0]
    switch cmd.(type) {
    case string:
        _cmd := cmd.(string)
        if strings.Contains(_cmd, "?") {
            sc := &subcond{
                command: cmd.(string),
                values:  make([]interface{}, 0),
            }
            if len(opts) > 1 {
                sc.values = append(sc.values, opts[1:]...)
            }
            c.conds = append(c.conds, sc)
        } else {
            c.conds = append(c.conds, _cmd)
            if len(opts) > 1 {
                c.Add(opts[1:]...)
            }
        }
    case []interface{}:
        c.Add(cmd.([]interface{})...)
        if len(opts) > 1 {
            c.Add(opts[1:]...)
        }
    case *subcond:
        c.conds = append(c.conds, cmd.(*subcond))
        if len(opts) > 1 {
            c.Add(opts[1:]...)
        }
    case *Condition:
        c.conds = append(c.conds, cmd.(*Condition))
        if len(opts) > 1 {
            c.Add(opts[1:]...)
        }
    case map[string]interface{}:
        c.addMap(cmd.(map[string]interface{}))
        if len(opts) > 1 {
            c.Add(opts[1:]...)
        }
    }
    return
}

// Match 进行关键字匹配查询
func (c *Condition) Match(field, keyword string) {
    field = strings.TrimSpace(field)
    keyword = strings.TrimSpace(keyword)
    if field == "" || keyword == "" {
        return
    }
    pattern, err := regexp.Compile(keywordPattern)
    if err != nil {
        c.AddError(err)
        return
    }

    // 分词
    keywords := pattern.Split(keyword, -1)
    newKeywords := make([]string, 0)
    for _, kwd := range keywords {
        kwd = strings.TrimSpace(kwd)
        if kwd == "" {
            continue
        }
        newKeywords = append(newKeywords, kwd)
    }
    if len(newKeywords) == 0 {
        return
    }
    // 构造条件
    cond := New(WithOrLogic())
    for _, kwd := range newKeywords {
        cond.Add(fmt.Sprintf("%s LIKE '%%%s%%'", field, kwd))
    }
    c.Add(cond)
}

// MultiMatch 对多字段进行关键字匹配查询
func (c *Condition) MultiMatch(fields []string, keyword string) {
    if len(fields) == 0 {
        return
    }
    cond := New(WithOrLogic())
    for _, field := range fields {
        cond.Match(field, keyword)
    }
    c.Add(cond)
}

// addMap 将map转换为条件
func (c *Condition) addMap(m map[string]interface{}) {
    if len(m) == 0 {
        return
    }
    for k, v := range m {
        k = strings.TrimSpace(k)

        // 本身是一个逻辑分组
        uk := strings.TrimSpace(strings.ToUpper(k))
        if uk == LogicAnd || uk == LogicOr {
            cond := New(WithLogic(uk))
            cond.Add(v)
            c.Add(cond)
            continue
        }

        // 特殊的条件
        if strings.Contains(k, "?") {
            c.Add(k, v)
            continue
        }

        // 普通条件
        field := k
        op := "="
        pos := strings.Index(k, " ")
        if pos > 0 {
            field = k[:pos]
            op = k[pos+1:]
        }
        cond, err := c.buildWhere(field, op, v)
        if err != nil {
            c.AddError(err)
            break
        }
        c.Add(cond)
    }
    return
}

// buildWhere 构造where条件
func (c *Condition) buildWhere(field, op string, value interface{}) (*subcond, error) {
    cond := &subcond{
        command: "",
        values:  make([]interface{}, 0),
    }
    op = strings.ToUpper(strings.TrimSpace(op))
    if op == "" {
        op = "="
    }
    field = strings.ReplaceAll(field, "`", "")
    switch op {
    case "=", "!=", ">", ">=", "<", "<=", "<>", "LIKE", "NOT LIKE", "IS", "IN", "NOT IN":
        cond.command = fmt.Sprintf("%s %s ?", field, op)
        cond.values = append(cond.values, value)
    case "BETWEEN", "NOT BETWEEN":
        vals := c.parseValue2Array(value)
        if len(vals) != 2 {
            return nil, ErrInvalidCondition
        }
        cond.command = fmt.Sprintf("%s %s ? AND ?", field, op)
        cond.values = append(cond.values, vals...)
    default:
        return nil, ErrNotSupportedQuery
    }
    return cond, nil
}

// parseValue2Array 将值转换成数组
func (c *Condition) parseValue2Array(value interface{}) []interface{} {
    return goutil.SVal(value).Interface()
}

// MarkUselessQuery 标记为空查询，即当前条件查询结果为空
func (c *Condition) MarkUselessQuery() {
    c.Error = ErrUselessQuery
}

// Abort 基于已知条件可以明确知道查询为空，调用此接口终止后续逻辑
// 此方法与MarkUselessQuery相同
func (c *Condition) Abort() {
    c.Error = ErrUselessQuery
}

// AddError 添加错误信息
func (c *Condition) AddError(err error) {
    c.Error = err
}

// Size 获取子条件数量，用于判断条件是否为空
func (c *Condition) Size() int {
    return len(c.conds)
}

// Build 构造条件, 将查询条件与对应的值分别返回
func (c *Condition) Build() (string, []interface{}) {
    if len(c.conds) == 0 || c.Error != nil {
        return "1 != 1", []interface{}{}
    }
    logic := c.logic
    if logic != LogicAnd && logic != LogicOr {
        logic = LogicAnd
    }
    command := bytes.Buffer{}
    vals := make([]interface{}, 0)
    for i, cond := range c.conds {
        if i > 0 {
            command.WriteString(logic)
        }
        innerCmd := ""
        innerVals := make([]interface{}, 0)
        switch cond.(type) {
        case string:
            innerCmd = cond.(string)
        case *subcond:
            sc := cond.(*subcond)
            innerCmd = sc.command
            if len(sc.values) > 0 {
                innerVals = append(innerVals, sc.values...)
            }
        case *Condition:
            innerCmd, innerVals = cond.(*Condition).Build()
        default:
            // 暂未实现其它类型，故暂且跳过
            continue
        }

        command.WriteString(" (")
        command.WriteString(innerCmd)
        command.WriteString(") ")
        if innerVals != nil && len(innerVals) > 0 {
            vals = append(vals, innerVals...)
        }
    }
    return command.String(), vals
}

// Conditions 构造查询条件，将构造结果放在切片中返回，方便在gorm中直接使用
func (c *Condition) Conditions() []interface{} {
    if c.Size() == 0 {
        return nil
    }
    cmd, vals := c.Build()
    conds := make([]interface{}, 0)
    conds = append(conds, cmd)
    conds = append(conds, vals...)
    return conds
}

// String 返回条件的字面形式，只用于调试，不可作为查询条件使用
func (c *Condition) String() string {
    cmd, vals := c.Build()
    cmd = strings.ReplaceAll(cmd, "%", "%%")
    cmd = strings.ReplaceAll(cmd, "?", "%v")
    return fmt.Sprintf(cmd, vals...)
}
