package main

import (
	"fmt"
	"net/http"

	http_context "github.com/eolinker/goku-eosc/node/http-context"
)

func main() {
	//a:="ab="
	//i:=strings.Index(a,"=")
	//fmt.Println(a[:i])
	//fmt.Println(a[i+1:])
	a := "*"
	fmt.Println(a[1:])
	err := http.ListenAndServeTLS(":8181", "eolinker.csr", "eolinker.key", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := http_context.NewContext(r, w)
		ctx.ProxyRequest.Headers()
	}))
	fmt.Println(err)
}
