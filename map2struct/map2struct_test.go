package map2struct

import (
    "fmt"
    "testing"

    "github.com/whencome/goutil"
)

type Student struct {
    Name    string  `map2struct:"name"`
    Age     int     `map2struct:"age"`
    Score   float32 `map2struct:"score"`
    Address string  `map2struct:"-"`
}

var s = &Student{
    Name:    "Jack Smith",
    Age:     18,
    Score:   78.50,
    Address: "A secret address that can not be known for now",
}

var m = goutil.M{
    "name":    "Jack Smith",
    "age":     18,
    "score":   89.5,
    "address": "A secret address that can not be known for now",
}

func TestMarshalTag(t *testing.T) {
    s1 := &Student{}
    err := ToStructByTag(s1, m, "map2struct")
    if err != nil {
        fmt.Printf("marshal failed: %s\n", err)
        t.Fail()
    }
    fmt.Printf("student: %+v\n", s1)
    if s1.Name != m.GetString("name") || s1.Age != m.GetInt("age") || s1.Address == m.GetString("address") {
        fmt.Printf("marshal fail")
        t.Fail()
    }
}

func TestMarshal(t *testing.T) {
    s1 := &Student{}
    err := ToStruct(s1, m)
    if err != nil {
        fmt.Printf("marshal failed: %s\n", err)
        t.Fail()
    }
    fmt.Printf("student: %+v\n", s1)
    if s1.Name != m.GetString("name") || s1.Age != m.GetInt("age") || s1.Address == m.GetString("address") {
        fmt.Printf("marshal fail")
        t.Fail()
    }
}

func TestUnmarshalTag(t *testing.T) {
    m := ToMapByTag(s, "map2struct")
    fmt.Printf("map: %+v\n", m)
    if s.Name != m.GetString("name") || s.Age != m.GetInt("age") || m.Contains("address") {
        fmt.Printf("unmarshal fail")
        t.Fail()
    }
}

func TestUnmarshal(t *testing.T) {
    m := ToMap(s)
    fmt.Printf("map: %+v\n", m)
    if s.Name != m.GetString("name") || s.Age != m.GetInt("age") || m.Contains("address") {
        fmt.Printf("unmarshal fail")
        t.Fail()
    }
}
