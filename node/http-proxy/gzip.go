package http_proxy

import (
	"compress/gzip"
	"io"
	"io/ioutil"
)

func ParseGzip(b io.Reader) ([]byte, error) {
	r, err := gzip.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return data, nil
}
