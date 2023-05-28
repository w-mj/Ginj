package lib

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"strings"
)

type GinjInstance struct {
	engin *gin.Engine
}

func New(engine *gin.Engine) *GinjInstance {
	return &GinjInstance{
		engin: engine,
	}
}

// Handle /**
/**
annotation: GET /admin/data?UserName&Page=val:positive&User=map:mapFunc
*/
func (ins *GinjInstance) Handle(annotation string, handler any) {
	space := strings.Index(annotation, " ")
	if space <= 0 {
		panic("Illegal annotation: " + annotation)
	}
	method := annotation[0:space]
	questionMark := strings.Index(annotation, "?")
	if questionMark == -1 {
		questionMark = len(annotation)
	}

	path := annotation[space+1 : questionMark]
	query := annotation[questionMark+1:]
	queryList := strings.Split(query, "&")
	queryListIndex := 0

	inputs := CreateHandlerInputs()
	handTyp := reflect.TypeOf(handler)
	for i := 0; i < handTyp.NumIn(); i++ {
		in := handTyp.In(i)
		if in.String() == reflect.TypeOf(&gin.Context{}).String() {
			inputs.AddInput(in.String(), in, "")
		} else {
			tag := strings.TrimSpace(queryList[queryListIndex])
			equalMark := strings.Index(tag, "=")
			if equalMark < 0 {
				inputs.AddInput(tag, in, "")
			} else {
				inputs.AddInput(tag[0:equalMark], in, tag[equalMark+1:])
			}
			queryListIndex++
		}
	}
	ins.engin.Handle(method, path, wrapperFunction(inputs, handler))
}

func processHandlerOuts(c *gin.Context, out []reflect.Value) {
	if len(out) == 0 {
		return
	}

	if len(out) == 1 {
		if err, ok := out[0].Interface().(error); ok {
			if out[0].Interface() == nil {
				c.String(http.StatusOK, "")
			} else {
				c.String(http.StatusBadRequest, err.Error())
			}
			return
		} else {
			fmt.Println("Controller handler return only 1 must be error.")
			return
		}
	}
	var ret = out[0].Interface()
	if !out[1].IsNil() {
		c.String(http.StatusBadRequest, out[1].Interface().(error).Error())
	} else {
		if out[0].Type().Name() == "string" {
			c.String(http.StatusOK, ret.(string))
		} else {
			c.JSON(http.StatusOK, ret)
		}
	}
}

func wrapperFunction(inputList *HandlerInputList, handler any) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		ins := inputList.BuildInputValues(ctx)
		outs := reflect.ValueOf(handler).Call(ins)
		processHandlerOuts(ctx, outs)
	}
}

type annoHandler struct {
	Anno    string
	Handler any
}

var annoHandlerList []annoHandler

func AddAnnotatedRoute(anno string, handler any) {
	annoHandlerList = append(annoHandlerList, annoHandler{
		Anno:    anno,
		Handler: handler,
	})
}
