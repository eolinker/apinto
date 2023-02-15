package flag

import "flag"

var (
	ListenPort    = 9001
	BindIP        = ""
	ConfigAddress = "http://127.0.0.1:8080"
	TlsKey        = ""
	TlsPem        = ""
)

func Parse() {
	flag.IntVar(&ListenPort, "p", 9001, "please provide listen port")
	flag.StringVar(&BindIP, "ip", "", "Please provide bind ip")
	flag.StringVar(&ConfigAddress, "configd", "http://127.0.0.1:8080", "configd address")
	flag.StringVar(&TlsKey, "key", "", "if tls,key is required")
	flag.StringVar(&TlsPem, "pem", "", "if tls,pem is required")
	flag.Parse()
}
