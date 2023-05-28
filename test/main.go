package main

import (
	"github.com/gin-gonic/gin"
	"github.com/w-mj/ginj"
	"sync"
)

func MyHandler(abc int, def string) (map[string]any, error) {
	ans := map[string]any{}
	ans["abc"] = abc
	ans["def"] = def
	return ans, nil
}

// @Ginj: GET /index2?abc&def
func AnnotateHandler(abc int, def string) (map[string]any, error) {
	return MyHandler(abc, def)
}

var wg = sync.WaitGroup{}

//go:generate ginj_gen .
func StartServer() {
	r := gin.Default()
	j := ginj.New(r)
	j.LoadAnnotatedRote()
	j.Route("GET /index?abc&def", MyHandler)

	err := r.Run("127.0.0.1:8729")
	if err != nil {
		panic(err)
	}
}
