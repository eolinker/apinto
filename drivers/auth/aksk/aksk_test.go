package aksk

import (
	"io"
	"net/http"
	"testing"

	"github.com/valyala/fasthttp"

	http_context "github.com/eolinker/goku/node/http-context"
)

var akskConfig = []AKSKConfig{{
	AK:     "4c897cfdfca60a59983adc2627942e7e",
	SK:     "6bb8eee91f88336dd95b88a66709f0a3286ce1abf73453acc4619bc142d64040",
	Labels: map[string]string{},
	Expire: 1658740726, //2022-07-25 17:18:46
}}

var testContexts = make([]*http_context.Context, 0, 10)

func TestAKSK(t *testing.T) {
	testAKSK := &aksk{
		id:             "123",
		name:           "name",
		hideCredential: true,
		users:          &akskUsers{users: akskConfig},
	}

	createTestContext()

	err := testAKSK.Auth(testContexts[0])
	//if err != nil {
	//	t.Errorf("测试1：预期是能够通过鉴权，结果是%s", err.Error())
	//}

	err = testAKSK.Auth(testContexts[1])
	if err == nil {
		t.Errorf("测试2：预期是不能够通过鉴权%s:，结果是nil", err.Error())
	}

	err = testAKSK.Auth(testContexts[2])
	if err == nil {
		t.Errorf("测试3：预期是不能够通过鉴权%s:，结果是nil", err.Error())
	}
}

func createTestContext() {
	//使用正确sk加密后的签名

	// http
	//request1, _ := http.NewRequest("GET", "http://www.demo.com/demo/login?parm1=value1&parm2=", &body{})
	//request1.Header.SetDriver("Authorization-Type", "ak/sk")
	//request1.Header.SetDriver("Content-Type", "application/json")
	//request1.Header.SetDriver("x-gateway-date", "20200605T104456Z")
	//request1.Header.SetDriver("Authorization", "HMAC-SHA256 Access=4c897cfdfca60a59983adc2627942e7e, SignedHeaders=content-type;host;x-gateway-date, Signature=0c3d2598d931f36ca7d261d52dcd29f09d6573671bd593b7cbc55f73eb942758")
	//Context1 := http_context.NewContext(request1, &writer{})

	// fast http
	request1 := fasthttp.AcquireRequest()
	request1.SetRequestURI("http://www.demo.com/demo/login?parm1=value1&parm2=")
	request1.Header.SetMethod(fasthttp.MethodGet)
	request1.Header.Set("Authorization-Type", "ak/sk")
	request1.Header.Set("Content-Type", "application/json")
	request1.Header.Set("x-gateway-date", "20200605T104456Z")
	request1.Header.Set("Authorization", "HMAC-SHA256 Access=4c897cfdfca60a59983adc2627942e7e, SignedHeaders=content-type;host;x-gateway-date, Signature=0c3d2598d931f36ca7d261d52dcd29f09d6573671bd593b7cbc55f73eb942758")
	context1 := &fasthttp.RequestCtx{
		Request:  *fasthttp.AcquireRequest(),
		Response: *fasthttp.AcquireResponse(),
	}
	request1.CopyTo(&context1.Request)

	Context1 := http_context.NewContext(context1)

	testContexts = append(testContexts, Context1)

	//使用错误sk加密后的签名

	// http
	//request2, _ := http.NewRequest("GET", "http://www.demo.com/demo/login?parm1=value1&parm2=", &body{})
	//request2.Header.SetDriver("Authorization-Type", "ak/sk")
	//request2.Header.SetDriver("Content-Type", "application/json")
	//request2.Header.SetDriver("x-gateway-date", "20200605T104456Z")
	//request2.Header.SetDriver("Authorization", "HMAC-SHA256 Access=4c897cfdfca60a59983adc2627942e7e, SignedHeaders=content-type;host;x-gateway-date, Signature=bb18110ddf327a9c1222a551527896d59cb854ca9084078cfa3a6eb23de3ddb8")
	//Context2 := http_context.NewContext(request2, &writer{})

	// https
	request2 := fasthttp.AcquireRequest()
	request2.SetRequestURI("http://www.demo.com/demo/login?parm1=value1&parm2=")
	request2.Header.SetMethod(fasthttp.MethodGet)
	request2.Header.Set("Authorization-Type", "ak/sk")
	request2.Header.Set("Content-Type", "application/json")
	request2.Header.Set("x-gateway-date", "20200605T104456Z")
	request2.Header.Set("Authorization", "HMAC-SHA256 Access=4c897cfdfca60a59983adc2627942e7e, SignedHeaders=content-type;host;x-gateway-date, Signature=bb18110ddf327a9c1222a551527896d59cb854ca9084078cfa3a6eb23de3ddb8")
	context2 := &fasthttp.RequestCtx{
		Request:  *fasthttp.AcquireRequest(),
		Response: *fasthttp.AcquireResponse(),
	}
	request2.CopyTo(&context2.Request)

	Context2 := http_context.NewContext(context2)
	testContexts = append(testContexts, Context2)

	//传输了不存在的ak
	// http
	//request3, _ := http.NewRequest("GET", "http://www.demo.com/demo/login?parm1=value1&parm2=", &body{})
	//request3.Header.SetDriver("Authorization-Type", "ak/sk")
	//request3.Header.SetDriver("Content-Type", "application/json")
	//request3.Header.SetDriver("x-gateway-date", "20200605T104456Z")
	//request3.Header.SetDriver("Authorization", "HMAC-SHA256 Access=dsaasdasda, SignedHeaders=content-type;host;x-gateway-date, Signature=0c3d2598d931f36ca7d261d52dcd29f09d6573671bd593b7cbc55f73eb942758")
	//Context3 := http_context.NewContext(request3, &writer{})
	//testContexts = append(testContexts, Context3)

	// fast http
	request3 := fasthttp.AcquireRequest()
	request3.SetRequestURI("http://www.demo.com/demo/login?parm1=value1&parm2=")
	request3.Header.SetMethod(fasthttp.MethodGet)
	request3.Header.Set("Authorization-Type", "ak/sk")
	request3.Header.Set("Content-Type", "application/json")
	request3.Header.Set("x-gateway-date", "20200605T104456Z")
	request3.Header.Set("Authorization", "HMAC-SHA256 Access=dsaasdasda, SignedHeaders=content-type;host;x-gateway-date, Signature=0c3d2598d931f36ca7d261d52dcd29f09d6573671bd593b7cbc55f73eb942758")
	context3 := &fasthttp.RequestCtx{
		Request:  *fasthttp.AcquireRequest(),
		Response: *fasthttp.AcquireResponse(),
	}
	request3.CopyTo(&context3.Request)
	Context3 := http_context.NewContext(context3)
	testContexts = append(testContexts, Context3)
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
