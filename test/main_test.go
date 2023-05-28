package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
	"time"
)

var url = "http://127.0.0.1:8729"

var a *assert.Assertions

func get(path string) *http.Response {
	request, err := http.NewRequest("GET", url+path, nil)
	a.NoError(err)
	client := http.Client{}
	do, err := client.Do(request)
	a.NoError(err)
	return do
}

func getJson(r *http.Response) map[string]any {
	all, err := io.ReadAll(r.Body)
	a.NoError(err)
	ans := map[string]any{}
	err = json.Unmarshal(all, &ans)
	a.NoError(err)
	return ans
}

func TestMainFunc(m *testing.T) {
	a = assert.New(m)
	go StartServer()
	time.Sleep(time.Second * 2)
	ans := get("/index2?abc=123&def=hello")
	a.Equal(ans.StatusCode, http.StatusOK)
	j := getJson(ans)
	a.Equal(j["abc"].(float64), float64(123))
	a.Equal(j["def"].(string), "hello")
}
