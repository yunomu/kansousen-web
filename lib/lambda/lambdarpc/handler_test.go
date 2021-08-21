package lambdarpc

import (
	"testing"

	"context"
	"reflect"

	"google.golang.org/protobuf/proto"
)

type testSvc struct{}

func (*testSvc) ValidMethod(context.Context, *Context, proto.Message) (proto.Message, error) {
	return nil, nil
}

func TestValidMethod_valid(t *testing.T) {
	svc := &testSvc{}
	svcType := reflect.TypeOf(svc)
	method, ok := svcType.MethodByName("ValidMethod")
	if !ok {
		t.Fatalf("ValidMethod not found")
	}

	if err := validMethod(svcType, method); err != nil {
		if e, ok := err.(*ErrInvalidMethod); !ok {
			t.Errorf("validMethod unknown error: %v", err)
		} else {
			t.Errorf("InvalidMethod: %s value=`%v`", e.details, e.errParam)
		}
	}
}
