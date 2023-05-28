package ginj

import (
	"github.com/gin-gonic/gin"
	"github.com/w-mj/ginj/lib"
)

func New(r *gin.Engine) *lib.GinjInstance {
	return lib.New(r)
}
