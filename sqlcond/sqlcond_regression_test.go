package sqlcond

import (
	"strings"
	"testing"
)

// TestMatchWithPlaceholder 测试关键字匹配使用占位符传参
func TestMatchWithPlaceholder(t *testing.T) {
	c := New()
	c.Match("name", "jack's dog")
	if c.Error != nil {
		t.Fatalf("Match failed: %v", c.Error)
	}
	cmd, vals := c.Build()
	if strings.Contains(cmd, "jack") {
		t.Fatalf("keyword should not be inlined into sql: %s", cmd)
	}
	if len(vals) != 2 {
		t.Fatalf("unexpected values: %v", vals)
	}
	// 多个关键字按OR连接
	if !strings.Contains(cmd, "OR") {
		t.Fatalf("keywords should be combined with OR: %s", cmd)
	}
	for _, v := range vals {
		if !strings.HasPrefix(v.(string), "%") || !strings.HasSuffix(v.(string), "%") {
			t.Fatalf("unexpected like value: %v", v)
		}
	}
}

// TestEmptyConditionGroup 测试空的子条件分组不会生成永假条件
func TestEmptyConditionGroup(t *testing.T) {
	c := New()
	// 空关键字的Match不会产生任何条件
	c.Match("name", "  ")
	c.MultiMatch([]string{"name", "title"}, "")
	c.Add(New())
	cmd, vals := c.Build()
	if cmd != "1 != 1" {
		t.Fatalf("empty condition should build 1 != 1, got %s", cmd)
	}
	if len(vals) != 0 {
		t.Fatalf("unexpected vals: %v", vals)
	}
	if c.IsOK() {
		t.Fatal("empty condition should not be ok")
	}

	// 有空条件混入时，有效条件仍然可用
	c2 := New()
	c2.Add(New())
	c2.Add("a = ?", 1)
	cmd, vals = c2.Build()
	if !strings.Contains(cmd, "a = ?") || len(vals) != 1 {
		t.Fatalf("valid condition lost: %s, %v", cmd, vals)
	}
}

// TestMapCondition 测试map方式的查询条件
func TestMapCondition(t *testing.T) {
	c := New()
	c.Add(map[string]interface{}{
		"name":   "jack",
		"age >":  18,
		"status": []string{"a", "b"},
	})
	if c.Error != nil {
		t.Fatalf("build condition failed: %v", c.Error)
	}
	cmd, vals := c.Build()
	if !strings.Contains(cmd, "name = ?") {
		t.Fatalf("unexpected sql: %s", cmd)
	}
	if !strings.Contains(cmd, "age > ?") {
		t.Fatalf("unexpected sql: %s", cmd)
	}
	if len(vals) != 3 {
		t.Fatalf("unexpected vals: %v", vals)
	}
}

// TestUselessQuery 测试空查询标记
func TestUselessQuery(t *testing.T) {
	inner := New()
	inner.MarkUselessQuery()

	// AND 逻辑下，子条件无用则整个条件无用
	outer := New()
	outer.Add(inner)
	if outer.Error != ErrUselessQuery {
		t.Fatalf("useless query should be propagated in AND logic, got %v", outer.Error)
	}

	// OR 逻辑下，无用子条件被忽略
	outerOr := New(WithOrLogic())
	outerOr.Add(inner)
	outerOr.Add("a = ?", 1)
	cmd, vals := outerOr.Build()
	if !strings.Contains(cmd, "a = ?") || len(vals) != 1 {
		t.Fatalf("OR logic should keep valid condition: %s, %v", cmd, vals)
	}
}

// TestInvalidCondition 测试非法的字段与操作符
func TestInvalidCondition(t *testing.T) {
	c := New()
	c.Add(map[string]interface{}{"name ~": "jack"})
	if c.Error != ErrNotSupportedQuery {
		t.Fatalf("unsupported operator should return error, got %v", c.Error)
	}
	cmd, _ := c.Build()
	if cmd != "1 != 1" {
		t.Fatalf("error condition should build 1 != 1, got %s", cmd)
	}
}
