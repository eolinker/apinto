package main

import (
	"github.com/valyala/fasthttp"
)

func main() {

	s := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {

			ctx.SetStatusCode(200)
			ctx.Write([]byte("ok"))
		},
	}
	s.ListenAndServe(":8082")
}
