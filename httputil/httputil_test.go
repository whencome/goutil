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

func TestDownloadFile(t *testing.T) {
	url := "https://p5.itc.cn/q_70/images01/20210608/289a9f603c9e4b76ab3ee69f20dacea3.jpeg"
	dstFile := "D:/download/1001.jpg"
	err := DownloadFile(url, dstFile)
	if err != nil {
		t.Fail()
	}
}