package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/w-mj/ginj"
)

// @Ginj: GET /index2?abc&def
func AnnotateHandler(abc int, def string) (map[string]any, error) {
	if abc < 100 {
		return nil, errors.New("abc must grater than 100")
	}
	ans := map[string]any{}
	ans["abc"] = abc
	ans["def"] = def
	return ans, nil
}

//go:generate ginj_gen .
func StartServer() {
	r := gin.New()
	j := ginj.New(r)
	j.LoadAnnotatedRote()

	err := r.Run("127.0.0.1:8729")
	if err != nil {
		panic(err)
	}
}
