package lambdarpc

import (
	"testing"

	"context"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/aws/aws-lambda-go/lambdacontext"

	kifupb "github.com/yunomu/kansousen/proto/kifu"
)

type testSvc struct{}

func (*testSvc) ValidMethod(ctx context.Context, req *kifupb.GetKifuRequest) (*kifupb.GetKifuResponse, error) {
	userId := ctx.Value(UserIdField)

	return &kifupb.GetKifuResponse{
		UserId: userId.(string),
		KifuId: req.KifuId,
	}, nil
}

func (*testSvc) ErrorMethod(ctx context.Context, req *kifupb.GetKifuRequest) (*kifupb.GetKifuResponse, error) {
	return nil, errors.New("test error")
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

func TestHandler(t *testing.T) {
	svc := &testSvc{}

	h := NewHandler(svc)
	if err := h.Init(); err != nil {
		t.Fatalf("handler.Init: %v", err)
	}

	if len(h.methods) != 2 {
		t.Fatalf("Valid method number: expected=2, actual=%d", len(h.methods))
	}

	lc := &lambdacontext.LambdaContext{
		AwsRequestID: "aws-request-id",
		ClientContext: lambdacontext.ClientContext{
			Custom: map[string]string{
				UserIdField:     "user-id-test",
				FunctionIdField: "ValidMethod",
			},
		},
	}
	ctx := lambdacontext.NewContext(context.Background(), lc)

	req := &kifupb.GetKifuRequest{
		KifuId: "kifu-id-test",
	}
	reqBs, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("proto.Marshal(req): %v", err)
	}

	resBs, err := h.Invoke(ctx, reqBs)
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}

	res := &kifupb.GetKifuResponse{}
	if err := json.Unmarshal(resBs, res); err != nil {
		t.Fatalf("proto.Unmarshal(res): %v", err)
	}

	if req.KifuId != res.KifuId {
		t.Errorf("kifuId is mismathc: req=%s res=%s", req.KifuId, res.KifuId)
	}
	if "user-id-test" != res.UserId {
		t.Errorf("UserId is mismathc: req=user-id-test res=%s", res.UserId)
	}
}
