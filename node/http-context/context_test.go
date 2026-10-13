package http_context

import (
	"bytes"
	"io"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestCopyUpstreamResponseFixedLengthStream(t *testing.T) {
	payload := []byte("fixed-length streamed body")
	src := fasthttp.AcquireResponse()
	src.SetStatusCode(fasthttp.StatusOK)
	src.Header.Set("X-Upstream", "fixed")
	src.SetBodyStream(bytes.NewReader(payload), len(payload))

	dst := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(dst)
	copyUpstreamResponse(src, dst)
	fasthttp.ReleaseResponse(src)

	if !bytes.Equal(dst.Body(), payload) {
		t.Fatalf("destination body = %q, want %q", dst.Body(), payload)
	}
	if dst.Header.ContentLength() != len(payload) {
		t.Fatalf("content length = %d, want %d", dst.Header.ContentLength(), len(payload))
	}
	if string(dst.Header.Peek("X-Upstream")) != "fixed" {
		t.Fatalf("header X-Upstream = %q, want fixed", dst.Header.Peek("X-Upstream"))
	}
	if dst.StatusCode() != fasthttp.StatusOK {
		t.Fatalf("status = %d, want %d", dst.StatusCode(), fasthttp.StatusOK)
	}
}

func TestCopyUpstreamResponseNonStream(t *testing.T) {
	payload := []byte("buffered body")
	src := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(src)
	src.SetStatusCode(fasthttp.StatusCreated)
	src.Header.Set("X-Upstream", "buffered")
	src.SetBody(payload)

	dst := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(dst)
	copyUpstreamResponse(src, dst)

	if !bytes.Equal(dst.Body(), payload) {
		t.Fatalf("destination body = %q, want %q", dst.Body(), payload)
	}
	if dst.StatusCode() != fasthttp.StatusCreated {
		t.Fatalf("status = %d, want %d", dst.StatusCode(), fasthttp.StatusCreated)
	}
	if string(dst.Header.Peek("X-Upstream")) != "buffered" {
		t.Fatalf("header X-Upstream = %q, want buffered", dst.Header.Peek("X-Upstream"))
	}
}

func TestCopyUpstreamResponseUnknownLengthStreamNotSwallowed(t *testing.T) {
	payload := []byte("chunked stream body")
	src := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(src)
	src.SetStatusCode(fasthttp.StatusOK)
	src.SetBodyStream(bytes.NewReader(payload), -1)

	dst := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(dst)
	copyUpstreamResponse(src, dst)

	if len(dst.Body()) != 0 {
		t.Fatalf("destination body = %q, want empty so the stream stays unread", dst.Body())
	}
	if !src.IsBodyStream() || src.Header.ContentLength() >= 0 {
		t.Fatalf("source stream was consumed: isStream=%v contentLength=%d", src.IsBodyStream(), src.Header.ContentLength())
	}
	got, err := io.ReadAll(src.BodyStream())
	if err != nil {
		t.Fatalf("read remaining stream: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("remaining stream = %q, want %q", got, payload)
	}
}
