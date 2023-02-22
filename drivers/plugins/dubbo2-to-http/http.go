package dubbo2_to_http

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

var client = http.DefaultClient

type Client struct {
	method string
	body   []byte
	path   string
}

func NewClient(method string, body []byte, path string) *Client {
	return &Client{method: method, body: body, path: path}
}

func (h *Client) dial(address string, timeout time.Duration) ([]byte, error) {

	client.Timeout = timeout
	request, err := http.NewRequest(h.method, address, bytes.NewReader(h.body))

	path := h.path
	if path != "" {
		if path[:1] != "/" {
			path = "/" + path
		}
		request.URL.Path = path
	}

	if err != nil {
		return nil, err
	}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return resBody, nil
}
