package main

import (
	"fmt"
	"ginj/lib"
	"github.com/gin-gonic/gin"
)

func MyHandler(abc int, def string) (map[string]any, error) {
	ans := map[string]any{}
	ans["abc"] = abc
	ans["def"] = def
	return ans, nil
}

// @Ginj: GET /index?abc&def
func AnnotateHandler(abc int, def string) (map[string]any, error) {
	return MyHandler(abc, def)
}

func main() {
	r := gin.Default()
	j := lib.New(r)
	j.Handle("GET /index?abc&def", MyHandler)
	err := r.Run(":8000")
	if err != nil {
		fmt.Println(err)
		return
	}
}
