package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
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
	client := &http.Client{}
	err := http.ListenAndServe(":8082", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "172.18.189.60"
		resp, err := client.Do(r)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		for key, value := range resp.Header {
			w.Header().Set(key, value[0])
		}

		w.Write(body)
	}))
	fmt.Println(err)
}
