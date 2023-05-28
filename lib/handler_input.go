package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type Validator struct {
	Name     string
	Validate func(any) bool
}

type InputPosition int

const (
	InputPositionNone    = 0x0
	InputPositionQuery   = 0x1
	InputPositionParam   = 0x2
	InputPositionContext = 0x4
	InputPositionPayload = 0x8
	InputPositionAny     = 0xf
)

type HandlerInput struct {
	Type       reflect.Type
	Name       string
	Map        func(string) any
	Validators []Validator
	Position   InputPosition
}

type HandlerInputList struct {
	Inputs []HandlerInput
}

func CreateHandlerInputs() *HandlerInputList {
	return &HandlerInputList{}
}

func (l *HandlerInputList) AddInput(name string, typ reflect.Type, tagStr string) {
	input := HandlerInput{
		Name: name,
		Type: typ,
	}
	tags := strings.Split(tagStr, ",")
	for _, t := range tags {
		if len(strings.TrimSpace(t)) == 0 {
			continue
		}
		switch t {
		case "payload":
			input.Position = InputPositionPayload
		default:
			fmt.Printf("Ginj.AddInput: Unknown Tag %s\n", t)
		}
	}
	if input.Position == InputPositionNone {
		input.Position = InputPositionAny
	}
	l.Inputs = append(l.Inputs, input)
}

func BuildHandlerInputsFromStructType(structType reflect.Type) *HandlerInputList {
	inputs := CreateHandlerInputs()
	fields := structType.NumField()
	for i := 0; i < fields; i++ {
		field := structType.Field(i)
		inputs.AddInput(field.Name, field.Type, field.Tag.Get("ginj"))
	}
	return inputs
}

func (l *HandlerInputList) BuildInputValues(c *gin.Context) []reflect.Value {
	var ans []reflect.Value
	for i := range l.Inputs {
		t := l.Inputs[i].Type
		value := reflect.Value{}
		if t.Kind() == reflect.Pointer {
			value.Set(reflect.New(t.Elem()))
		} else {
			value = reflect.New(t).Elem()
		}
		if value.Type().String() == reflect.TypeOf(&gin.Context{}).String() {
			value.Set(reflect.ValueOf(c))
		} else {
			getValueFromGinContext(c, &l.Inputs[i], value)
		}
		ans = append(ans, value)
	}
	return ans
}

func getValueFromGinContext(c *gin.Context, input *HandlerInput, value reflect.Value) {
	dataStr := ""
	contentType := c.Request.Header.Get("Content-Type")
	if input.Name != "" {
		if v, e := c.Get(input.Name); e && input.Position&InputPositionContext > 0 {
			dataStr = v.(string)
		}
		if v := c.Param(input.Name); v != "" && input.Position&InputPositionParam > 0 {
			dataStr = v
		}
		if v := c.Query(input.Name); v != "" && input.Position&InputPositionQuery > 0 {
			dataStr = v
		}
	}

	if dataStr == "" && input.Position&InputPositionPayload >= 0 {
		all, err := io.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Println("Ginj getValueFromGinContext io.ReadAll")
			return
		}
		dataStr = string(all)
	}
	if dataStr == "" {
		return
	}
	if input.Map != nil {
		value.Set(reflect.ValueOf(input.Map(dataStr)))
	} else {
		err := setValue(dataStr, value, contentType)
		if err != nil {
			fmt.Printf("Ginj getValueFromGinContext: %v\n", err)
		}
	}
}

func setValue(data string, value reflect.Value, contentType string) error {
	switch value.Kind() {
	case reflect.Bool:
		v, err := strconv.ParseBool(data)
		if err != nil {
			return err
		}
		value.SetBool(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(data, 10, 64)
		if err != nil {
			return err
		}
		value.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(data, 10, 64)
		if err != nil {
			return err
		}
		value.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(data, 64)
		if err != nil {
			return err
		}
		value.SetFloat(v)
	case reflect.Pointer:
		return setValue(data, reflect.Indirect(value), contentType)
	case reflect.String:
		value.SetString(data)
	case reflect.Struct:
		switch contentType {
		case "application/json":
			err := json.Unmarshal([]byte(data), value.Interface())
			if err != nil {
				return err
			}
		default:
			fmt.Printf("Ginj.SetValue unsupported content-type %s\n", contentType)
			return errors.New("unsupported type")
		}
	default:
		fmt.Printf("Ginj.setValue unsupported type %s\n", value.Type().Name())
		return errors.New("unsupported type")
	}
	return nil
}
