package goutil

import "testing"

func TestMValByPath(t *testing.T) {
	var mData = map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": 1,
			},
		},
		"d": 2,
		"e": map[string]interface{}{
			"f": map[string]interface{}{
				"g": 3,
			},
		},
	}
	mv := MVal(mData)
	subM, ok := mv.MValByPath("a.b")
	if !ok {
		t.Error("MValByPath error: path not found")
	}
	cVal := subM.GetInt("c")
	if cVal != 1 {
		t.Error("MValByPath error: value not match")
	}
	subM1, ok := mv.MValByPath("d.f")
	if ok || subM1 != nil {
		t.Error("MValByPath error: path not exists but found")
	}
}

func TestSMValByPath(t *testing.T) {
	var mData = map[string]interface{}{
		"a": map[string]interface{}{
			"b": []int{1, 2, 3},
		},
		"c": 2,
		"d": map[string]interface{}{
			"e": map[string]interface{}{
				"f": []map[string]interface{}{
					{"k1": "v1"},
					{"k2": "v2"},
					{"k3": "v3"},
				},
			},
		},
	}
	mv := MVal(mData)
	subSM, ok := mv.SMValByPath("a.b")
	if ok || subSM != nil {
		t.Error("SMValByPath failed: value type not match")
	}
	subSM, ok = mv.SMValByPath("d.e.f")
	if !ok || len(subSM) != 3 {
		t.Error("SMValByPath failed: got data fail")
	}
	if subSM[0]["k1"] != "v1" || subSM[1]["k2"] != "v2" || subSM[2]["k3"] != "v3" {
		t.Error("SMValByPath failed: got data fail")
	}
}
