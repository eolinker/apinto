package main

import (
	"fmt"
	"unsafe"
)

func main() {
	data := make([]byte, 8*1024*1024)
	for i := range data {
		data[i] = 1
	}
	fmt.Printf("n2 的类型 %T n2占中的大小是 %d G", data, unsafe.Sizeof(data))
	//s := &fasthttp.Server{
	//	Handler: func(ctx *fasthttp.RequestCtx) {
	//
	//		ctx.SetStatusCode(200)
	//		ctx.Write([]byte("ok"))
	//	},
	//}
	//s.ListenAndServe(":8082")
}
