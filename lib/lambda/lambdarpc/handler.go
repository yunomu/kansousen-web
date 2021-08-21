package lambdarpc

import (
	"context"
	"errors"
	"reflect"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

var (
	ErrNoLambdaContext   = errors.New("no lambda context")
	ErrNoClientContext   = errors.New("no client context")
	ErrUnauthenticated   = errors.New("unauthenticated")
	ErrNoMethodSpecified = errors.New("no method name in context")
	ErrMethodNotFound    = errors.New("method not found")
)

type ErrInvalidMethod struct {
	details  string
	errParam interface{}
}

func (e *ErrInvalidMethod) Error() string {
	return "invalid method signature: func (obj) (context.Context, *Context, proto.Message) (proto.Message, errors)"
}

type Handler struct {
	service interface{}

	serviceType reflect.Type
}

var _ lambda.Handler = (*Handler)(nil)

var (
	contextType        = reflect.TypeOf((*context.Context)(nil)).Elem()
	requestContextType = reflect.TypeOf((*Context)(nil))
	errorType          = reflect.TypeOf((*error)(nil)).Elem()
	messageType        = reflect.TypeOf((*proto.Message)(nil)).Elem()
)

func NewHandler(service interface{}) *Handler {
	return &Handler{
		service: service,

		serviceType: reflect.TypeOf(service),
	}
}

func validMethod(svcType reflect.Type, m reflect.Method) error {
	if m.Type.NumIn() != 4 {
		return &ErrInvalidMethod{
			details:  "invalid parameter number",
			errParam: m.Type.NumIn(),
		}
	}

	if m.Type.In(0).String() != svcType.String() {
		return &ErrInvalidMethod{
			details:  "invalid service type",
			errParam: m.Type.In(0),
		}
	}

	if m.Type.In(1).String() != contextType.String() {
		return &ErrInvalidMethod{
			details:  "1st param is not context.Context",
			errParam: m.Type.In(1),
		}
	}

	if m.Type.In(2).String() != requestContextType.String() {
		return &ErrInvalidMethod{
			details:  "2nd param is not Context",
			errParam: m.Type.In(2),
		}
	}

	if !m.Type.In(3).Implements(messageType) {
		return &ErrInvalidMethod{
			details:  "3rd param don't implement proto.Message",
			errParam: m.Type.In(3),
		}
	}

	if m.Type.NumOut() != 2 {
		return &ErrInvalidMethod{
			details:  "invalid number of return value",
			errParam: m.Type.NumOut(),
		}
	}

	if !m.Type.Out(0).Implements(messageType) {
		return &ErrInvalidMethod{
			details:  "1st return value don't implement proto.Message",
			errParam: m.Type.Out(0),
		}
	}

	if m.Type.Out(1).String() != errorType.String() {
		return &ErrInvalidMethod{
			details:  "2nd return value is not error",
			errParam: m.Type.Out(1),
		}
	}

	return nil
}

func (h *Handler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return nil, ErrNoLambdaContext
	}

	reqCtx := &Context{
		RequestId: lc.AwsRequestID,
	}

	custom := lc.ClientContext.Custom
	if custom == nil {
		return nil, ErrNoClientContext
	}

	reqCtx.ApiRequestId = custom[ApiRequestIdField]
	reqCtx.UserId, ok = custom[UserIdField]
	if !ok {
		return nil, ErrUnauthenticated
	}

	methodName, ok := custom[MethodField]
	if !ok {
		return nil, ErrNoMethodSpecified
	}

	m, ok := h.serviceType.MethodByName(methodName)
	if !ok {
		return nil, ErrMethodNotFound
	}

	if err := validMethod(h.serviceType, m); err != nil {
		return nil, err
	}

	reqType := m.Type.In(3)
	if reqType.Kind() == reflect.Ptr {
		reqType = reqType.Elem()
	}

	reqVal := reflect.New(reqType)
	reqMsg, ok := reqVal.Interface().(proto.Message)
	if !ok {
		panic("request value is not implement proto.Message")
	}

	if err := protojson.Unmarshal(payload, reqMsg); err != nil {
		return nil, err
	}

	rets := m.Func.Call([]reflect.Value{
		reflect.ValueOf(h.service),
		reflect.ValueOf(ctx),
		reflect.ValueOf(reqCtx),
		reqVal,
	})

	if len(rets) != 2 {
		panic("invalid number of return values")
	}

	err, ok := rets[1].Interface().(error)
	if !ok {
		panic("2nd return value is not error")
	}
	if err != nil {
		return nil, err
	}

	resMsg, ok := rets[0].Interface().(proto.Message)
	if !ok {
		panic("1st return value don't implement proto.Message")
	}

	bs, err := proto.Marshal(resMsg)
	if err != nil {
		return nil, err
	}

	return bs, nil
}
