package main

var requestFunc = []func([]string, map[string]string) error{
	CurrentRequest,
	StreamRequest,
	StreamResponse,
	AllStream,
}

func main() {
	Parse()
	md := map[string]string{
		"app": "apinto",
	}
	names := []string{
		"apinto",
		"eolink",
	}
	for _, f := range requestFunc {
		f(names, md)
	}
}
