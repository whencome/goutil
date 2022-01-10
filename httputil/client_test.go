package httputil

import (
	"io/ioutil"
	"testing"
)

func TestGet(t *testing.T) {
	url := "https://cn.bing.com/search"
	client := NewHttpClient()
	client.SetDebug(true)
	client.AddParam("q", "编程语言排行")
	resp, err := client.Get(url)
	if err != nil {
		t.Log("get err: ", err)
		t.Fail()
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Log("read err: ", err)
		t.Fail()
		return
	}
	t.Log(string(body))
}
