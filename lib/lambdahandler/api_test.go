package lambdahandler

import (
	"testing"

	"context"
)

func TestNewAPIHandler_AddAPIHandler(t *testing.T) {
	path := "/test"
	method := "GET"
	h := NewAPIHandler(AddAPIHandler(path, method, func(context.Context, *RequestContext, *Request) *Response {
		return nil
	}))

	p, ok := h.handlers[path]
	if !ok {
		t.Logf("path len=%d", len(h.handlers))
		for k, _ := range h.handlers {
			t.Logf("path[%s]", k)
		}
		t.Fatalf("path not found: path=%v", path)
	}

	handler, ok := p[method]
	if !ok {
		t.Fatalf("method not found: method=%v", method)
	}

	if handler == nil {
		t.Fatalf("handler not found")
	}
}
