package apihandler

import (
	"testing"

	"context"

	"google.golang.org/protobuf/proto"

	"github.com/yunomu/kansousen/lib/lambda/requestcontext"
)

func TestNewHandler_AddHandler(t *testing.T) {
	path := "/test"
	method := "GET"
	h := NewHandler(
		AddHandler(path, method, func(context.Context, *requestcontext.Context, *Request) (proto.Message, Error) {
			return nil, nil
		}),
	)

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
