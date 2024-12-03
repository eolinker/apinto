package loki

import (
	"bytes"
	"net/http"

	"github.com/eolinker/eosc"
)

type Output struct {
	url       string
	method    string
	headers   map[string]string
	formatter eosc.IFormatter
}

func (o *Output) Output(entry eosc.IEntry) error {

	return eosc.ErrorWorkerNotRunning
}

func (o *Output) genRequest(body []byte) (*http.Request, error) {
	req, err := http.NewRequest(o.method, o.url, bytes.NewReader(body))
	if err != nil {

	}
}
