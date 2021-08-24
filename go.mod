module github.com/eolinker/goku

go 1.15

require (
	github.com/eolinker/eosc v0.0.8
	github.com/eolinker/goku-plugin v0.1.4 // indirect
	github.com/eolinker/goku-standard-plugin v0.1.5
	github.com/ghodss/yaml v1.0.0
	github.com/go-basic/uuid v1.0.0
	github.com/hashicorp/consul/api v1.9.1
	github.com/nacos-group/nacos-sdk-go v1.0.8 // indirect
	github.com/pkg/sftp v1.13.2
	github.com/robertkrimen/otto v0.0.0-20210614181706-373ff5438452
	github.com/satori/go.uuid v1.2.0
	github.com/sbunce/bson v0.0.0-20181119052045-2aa5ebe749b2 // indirect
	github.com/valyala/fasthttp v1.28.0
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
)
replace (
	github.com/eolinker/eosc v0.0.8 => ../eosc
)

