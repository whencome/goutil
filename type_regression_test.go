package goutil

import (
	"testing"
)

func TestXSDelete(t *testing.T) {
	xs := NewXS(nil)
	xs.Add(1, 1, 2, "a")
	if xs.Size() != 4 {
		t.Fatalf("unexpected size: %d", xs.Size())
	}
	xs.Delete(1)
	if xs.Size() != 2 {
		t.Fatalf("Delete failed, size = %d, slice = %v", xs.Size(), xs.Slice())
	}
	for _, v := range xs.Slice() {
		if String(v) == "1" {
			t.Fatalf("element 1 should be removed: %v", xs.Slice())
		}
	}
	if xs.Contains(1) {
		t.Fatal("Contains should return false after delete")
	}
	// 删除不存在的元素不应panic
	xs.Delete("not-exists")
	if xs.Size() != 2 {
		t.Fatalf("size changed after deleting a non-exist element: %d", xs.Size())
	}
}

func TestXSUniqueOrder(t *testing.T) {
	xs := NewXS(nil)
	xs.Add("c", "a", "b", "a")
	got := xs.Unique()
	want := []string{"c", "a", "b"}
	if len(got) != len(want) {
		t.Fatalf("Unique got %v", got)
	}
	for i, v := range got {
		if String(v) != want[i] {
			t.Fatalf("Unique order mismatch: got %v, want %v", got, want)
		}
	}
}

func TestXSIntersectAndUnion(t *testing.T) {
	xs := NewXS(nil)
	xs.Add(1, 2, 3)
	inter := xs.Intersect(S{2, 3, 4})
	if inter.Size() != 2 {
		t.Fatalf("Intersect got %v", inter.Slice())
	}
	if !inter.Contains(2) || !inter.Contains(3) {
		t.Fatalf("Intersect result error: %v", inter.Slice())
	}
	union := xs.Union(S{3, 4})
	if union.Size() != 4 {
		t.Fatalf("Union got %v", union.Slice())
	}
}

func TestGetByPath(t *testing.T) {
	m := MVal(map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": 1,
			},
		},
		"d": 2,
	})
	if v, ok := m.GetByPath("a.b.c"); !ok || Int(v) != 1 {
		t.Fatalf("GetByPath failed: %v, %v", v, ok)
	}
	if v, ok := m.GetByPath("d"); !ok || Int(v) != 2 {
		t.Fatalf("GetByPath failed: %v, %v", v, ok)
	}
	if _, ok := m.GetByPath("d.e"); ok {
		t.Fatal("GetByPath should return false for a non-map intermediate value")
	}
	if _, ok := m.GetByPath("a.x.c"); ok {
		t.Fatal("GetByPath should return false for a non-exist path")
	}
	if _, ok := m.GetByPath(""); ok {
		t.Fatal("GetByPath should return false for empty path")
	}
}

// testTreeItem 测试用的树节点
type testTreeItem struct {
	key      string
	parent   string
	children []TreeItem
}

func (i *testTreeItem) Key() string           { return i.key }
func (i *testTreeItem) ParentKey() string     { return i.parent }
func (i *testTreeItem) AddChild(ti TreeItem)  { i.children = append(i.children, ti) }
func (i *testTreeItem) GetChild(k string) (TreeItem, bool) {
	for _, c := range i.children {
		if c.Key() == k {
			return c, true
		}
	}
	return nil, false
}

func TestTreeBuild(t *testing.T) {
	tree := NewTree()
	a := &testTreeItem{key: "a"}
	b := &testTreeItem{key: "b"}
	a1 := &testTreeItem{key: "a1", parent: "a"}
	a1x := &testTreeItem{key: "a1x", parent: "a1"}
	b1 := &testTreeItem{key: "b1", parent: "b"}
	// 故意打乱添加顺序，验证子节点能正确挂载
	tree.Add(a1x, b1, a1, a, b)

	roots := tree.Tree()
	if len(roots) != 2 {
		t.Fatalf("unexpected roots: %d", len(roots))
	}
	// 重复构建不应该产生重复的子节点
	roots = tree.Tree()
	c1, ok := roots[0].GetChild("a1")
	if !ok {
		t.Fatalf("child a1 not found in %s", roots[0].Key())
	}
	if _, ok := c1.GetChild("a1x"); !ok {
		t.Fatal("grandchild a1x not found")
	}
	if roots[0].(*testTreeItem).children == nil || len(roots[0].(*testTreeItem).children) != 1 {
		t.Fatalf("duplicated children after build twice: %d", len(roots[0].(*testTreeItem).children))
	}
}

func TestTreeCycle(t *testing.T) {
	tree := NewTree()
	// 构造 a -> b -> a 的环，确保不会死循环
	tree.Add(
		&testTreeItem{key: "a", parent: "b"},
		&testTreeItem{key: "b", parent: "a"},
	)
	roots := tree.Tree()
	if len(roots) != 0 {
		t.Fatalf("cyclic data should not be treated as a tree, got %d roots", len(roots))
	}
}
