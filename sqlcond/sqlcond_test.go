package sqlcond

import (
    "log"
    "testing"
)

func TestCondition_Build(t *testing.T) {
    cond := New()
    cond.Add("name = ? AND age > ?", "eric", 27)
    orCond := New(WithOrLogic())
    orCond.Add("class = ?", "A")
    orCond.Add("score > ?", 70)
    cond.Add(orCond)
    cond.Add("v1 = 1", "v2 = 2", "v3 = 3", "v4 = ?", 4)
    cmd, vals := cond.Build()
    log.Printf(cmd, vals)
}

func TestCondition_Match(t *testing.T) {
    cond := New()
    cond.Add("name = ? AND age > ?", "eric", 27)
    orCond := New(WithOrLogic())
    orCond.Add("class = ?", "A")
    orCond.Add("score > ?", 70)
    cond.Add(orCond)
    cond.Add("v1 = 1", "v2 = 2", "v3 = 3", "v4 = ?", 4)
    cond.Match("area", "%国际,城南.小区#")
    cmd, vals := cond.Build()
    log.Println(cmd, vals)
    log.Println(cond.String())
}
