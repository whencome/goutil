package array

import (
	"reflect"
	"testing"
	"time"
)

func TestChunk(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5}
	got := Chunk(arr, 2)
	want := [][]int{{1, 2}, {3, 4}, {5}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Chunk got %v, want %v", got, want)
	}
	// 只有一个分组
	if got := Chunk(arr, 10); len(got) != 1 || len(got[0]) != 5 {
		t.Fatalf("Chunk with a bigger size got %v", got)
	}
	// 历史实现中size为0或者负数会进入死循环，此处必须能够立即返回
	done := make(chan [][]int, 1)
	go func() {
		done <- Chunk(arr, 0)
	}()
	select {
	case got := <-done:
		if len(got) != 1 {
			t.Fatalf("Chunk with size 0 got %v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("Chunk with size 0 hangs")
	}
	if got := Chunk(arr, -1); len(got) != 1 {
		t.Fatalf("Chunk with negative size got %v", got)
	}
}

func TestContains(t *testing.T) {
	if !Contains([]string{"a", "b"}, "b") {
		t.Error("Contains should return true")
	}
	if Contains([]string{"a"}, "c") {
		t.Error("Contains should return false")
	}
	if Contains([]string{}, "a") {
		t.Error("Contains with empty slice should return false")
	}
}

func TestUniqueAndFilter(t *testing.T) {
	got := Unique([]int{1, 2, 2, 3, 1})
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("Unique got %v", got)
	}
	gotStr := Filter([]string{"a", "", "b"})
	if !reflect.DeepEqual(gotStr, []string{"a", "b"}) {
		t.Fatalf("Filter got %v", gotStr)
	}
	gotStr = UniqueAndFilter([]string{"a", "", "a", "b"})
	if !reflect.DeepEqual(gotStr, []string{"a", "b"}) {
		t.Fatalf("UniqueAndFilter got %v", gotStr)
	}
}

func TestRemove(t *testing.T) {
	got := Remove([]int{1, 2, 3, 2}, 2)
	if !reflect.DeepEqual(got, []int{1, 3}) {
		t.Fatalf("Remove got %v", got)
	}
	got = RemoveBatch([]int{1, 2, 3, 4}, 2, 4)
	if !reflect.DeepEqual(got, []int{1, 3}) {
		t.Fatalf("RemoveBatch got %v", got)
	}
	// 不传移除元素时应返回原切片
	src := []int{1, 2}
	if got := RemoveBatch(src, []int{}...); !reflect.DeepEqual(got, src) {
		t.Fatalf("RemoveBatch without elements got %v", got)
	}
}
