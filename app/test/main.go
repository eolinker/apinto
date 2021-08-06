package main

import (
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
)

func main() {
	//a:="ab="
	//i:=strings.Index(a,"=")
	//fmt.Println(a[:i])
	//fmt.Println(a[i+1:])
	//a := "*"
	//fmt.Println(a[1:])
	//err := http.ListenAndServeTLS(":8181", "eolinker.csr", "eolinker.key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//
	//	ctx := http_context.NewContext(r, w)
	//	ctx.ProxyRequest.Headers()
	//}))
	//fmt.Println(err)
	//transport := &http.Transport{TLSClientConfig: &tls.Config{
	//	InsecureSkipVerify: false,
	//},
	//	DialContext: (&net.Dialer{
	//		Timeout:   30 * time.Second, // 连接超时时间
	//		KeepAlive: 60 * time.Second, // 保持长连接的时间
	//	}).DialContext, // 设置连接的参数
	//	MaxIdleConns:          500,              // 最大空闲连接
	//	IdleConnTimeout:       60 * time.Second, // 空闲连接的超时时间
	//	ExpectContinueTimeout: 30 * time.Second, // 等待服务第一个响应的超时时间
	//	MaxIdleConnsPerHost:   100,              // 每个host保持的空闲连接数
	//}
	//client := &http.Client{Transport: transport}
	client := &fasthttp.Client{ReadTimeout: 30 * time.Second, MaxConnsPerHost: 4000}
	s := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			req := fasthttp.AcquireRequest()
			req.Header.SetMethod("POST")
			req.SetRequestURI("http://47.95.203.198:8080/Web/Test/params/print")
			var resp fasthttp.Response
			err := client.Do(req, &resp)
			//status, resp, err := fasthttp.Get(nil, "http://172.18.189.60/")
			if err != nil {
				fmt.Println("请求失败:", err.Error())
				return
			}

			if resp.StatusCode() != fasthttp.StatusOK {
				fmt.Println("请求没有成功:", resp.StatusCode())
				return
			}
			ctx.SetStatusCode(resp.StatusCode())
			ctx.Write(resp.Body())
		},
	}
	s.ListenAndServe(":8082")
}
