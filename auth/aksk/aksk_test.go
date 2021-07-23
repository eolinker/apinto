package aksk

import (
	http_context "github.com/eolinker/goku-eosc/node/http-context"
	"io"
	"log"
	"net/http"
	"testing"
)

var akskConfig = map[string]AKSKConfig{
	"4c897cfdfca60a59983adc2627942e7e": {
		AK:             "4c897cfdfca60a59983adc2627942e7e",
		HideCredential: true,
		Labels:         map[string]string{},
		Expire:         1658740726, //2022-07-25 17:18:46
	},
}

func TestAKSK(t *testing.T) {
	//aksk := &aksk{
	//	id:         "123",
	//	name:       "name",
	//	akskConfig: akskConfig,
	//}
	request, _ := http.NewRequest("GET", "http://www.demo.com/demo/login?parm1=value1&parm2=", &body{})
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-gateway-date", "20200605T104456Z")
	authorization := "Authorization: HMAC-SHA256 Access=4c897cfdfca60a59983adc2627942e7e, SignedHeaders=content-type;x-gateway-date, Signature=123123"
	oldContext := http_context.NewContext(request, &writer{})
	parseAuthorization(oldContext)
	log.Println(oldContext)
}

type body struct {
}

func (b body) Read(p []byte) (n int, err error) {
	return len(p), io.EOF
}

type writer struct {
}

func (w *writer) Header() http.Header {
	header := http.Header{}
	return header
}

func (w *writer) Write(bytes []byte) (int, error) {
	return len(bytes), nil
}

func (w *writer) WriteHeader(statusCode int) {
	return
}
