package main

import (
	"context"
	"errors"
	"fmt"
	. "gin-trace"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"time"
)

type Req struct {
	Cursor int64 `json:"cursor" binding:"required"`
}

type mError struct {
	Code    int
	Message string
}

func (m *mError) Error() string {
	return m.Message
}

func main() {
	app := gin.New()

	app.POST("/list", NewTrace(nil), func(c *gin.Context) {
		c.Bind(&gin.H{})
		c.JSON(http.StatusOK, gin.H{
			"A": "B",
		})
	})
	app.POST("/test", NewTrace(nil), Name(func(ctx context.Context, header *Header, req *Req) (interface{}, error) {
		//return gin.H{"A": "B"}, nil
		<-time.After(time.Second * 15)
		return map[string]interface{}{
			"A": "b",
		}, nil
	}))
	app.Run(":8080")
}

var _ctx context.Context = context.TODO()
var typeCtx = reflect.ValueOf(context.TODO())

type Header struct {
	Token string `json:"token" header:"token" binding:"required"`
}

var typeHeader = reflect.ValueOf(&Header{}).Type()

func Name(handler interface{}) gin.HandlerFunc {
	v := reflect.ValueOf(handler)
	if v.Kind() != reflect.Func {
		panic("input handler is not func")
	}
	ctx, header, req, err := GenIns(v)
	if err != nil {
		panic(err)
	}
	return func(c *gin.Context) {
		context, cancel := context.WithTimeout(context.TODO(), time.Second*5)
		defer cancel()
		ctx = reflect.ValueOf(context)
		if err := c.ShouldBindHeader(header.Interface()); err != nil {
			fmt.Println(err)
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if err := c.ShouldBind(req.Interface()); err != nil {
			fmt.Println(err)
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		var rsp interface{}
		var err error

		ch := make(chan interface{}, 1)
		call := func() {
			outs := v.Call([]reflect.Value{ctx, header, req})
			rsp, err = genOuts(outs)
			ch <- nil
		}
		go call()
		select {
		case <-ch:
		case <-context.Done():
			err = errors.New("request timeout")
		case <-time.After(time.Second * 5):
			err = errors.New("timeout")
		}
		if err != nil {
			fmt.Println(err)
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		c.JSON(http.StatusOK, rsp)
	}
}

func GenIns(v reflect.Value) (ctx, header, req reflect.Value, err error) {
	n := v.Type().NumIn()
	if n != 3 {
		panic("handler' in params is not 3")
	}

	ctxType := v.Type().In(0)
	if ctxType.AssignableTo(typeCtx.Type()) {
		err = errors.New("1nd param is not context.Context")
		return
	}
	ctx = reflect.New(typeCtx.Elem().Type())

	headerType := v.Type().In(1)
	if headerType.Kind() != reflect.Ptr {
		err = errors.New("2nd param is not point struct")
		return
	}
	if headerType != typeHeader {
		err = errors.New("2nd params is not *Header")
		return
	}

	header = reflect.New(headerType.Elem())

	reqType := v.Type().In(2)
	if reqType.Kind() != reflect.Ptr {
		err = errors.New("3nd param is not point")
		return
	}
	if reqType.Elem().Kind() != reflect.Struct {
		err = errors.New("3nd param is not a struct")
		return
	}
	req = reflect.New(reqType.Elem())
	//
	//for i := 0; i < n; i++ {
	//	in := v.Type().In(i)
	//	if in.Name() == "Context" {
	//		ctx = reflect.ValueOf(context.TODO())
	//	} else if in.Kind() == reflect.Ptr {
	//		in = in.Elem()
	//		if in.AssignableTo(typeHeader) {
	//			header = reflect.New(in)
	//		} else {
	//			req = reflect.New(in)
	//		}
	//	} else {
	//		fmt.Println(in.Name(), reflect.ValueOf(context.TODO()).Type().Name(), reflect.TypeOf(&Header{}).Kind())
	//		return ctx, header, req, fmt.Errorf("%dnd params is not fit", i+1)
	//	}
	//}
	return
}

func genOuts(outs []reflect.Value) (rsp interface{}, err error) {
	if len(outs) != 2 {
		return nil, errors.New("handler rsp count is not 2")
	}
	for _, out := range outs {
		if out.Type().Name() == "error" {
			if out.Interface() == nil {
				err = nil
			} else {
				err = out.Interface().(error)
			}
			//} else if out.Type().AssignableTo(reflect.TypeOf(map[string]interface{}{})) {
			//	fmt.Println("MAP")
			//	rsp = out.Interface()
		} else {
			rsp = out.Interface()
			fmt.Println(rsp)
		}
	}
	return
}
