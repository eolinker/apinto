package main

import (
	"fmt"
	"strings"
)

func main() {
	a:="ab="
	i:=strings.Index(a,"=")
	fmt.Println(a[:i])
	fmt.Println(a[i+1:])
}
