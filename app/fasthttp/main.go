package main

import (
	"io/ioutil"
	"time"

	"github.com/valyala/fasthttp"
)

func main() {
	data, err := ioutil.ReadFile("index.html")
	if err != nil {
		panic(err)
	}
	s := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			time.Sleep(10 * time.Millisecond)
			ctx.SetStatusCode(200)
			ctx.Write(data)
		},
	}
	s.ListenAndServe(":8082")
}
