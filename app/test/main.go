package main

import (
	"fmt"
	"sort"
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
	demo := []int{4, 1, 5, 2, 5, 15}
	sort.Sort(IntSlice(demo))
	fmt.Println(demo)
}

type IntSlice []int

func (a IntSlice) Len() int {
	return len(a)
}

func (a IntSlice) Less(i, j int) bool {
	return a[i] < a[j]
}

func (a IntSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
