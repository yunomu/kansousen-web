package lambdarpc

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

var (
	ErrNoLambdaContext     = errors.New("no lambda context")
	ErrNoClientContext     = errors.New("no client context")
	ErrUnauthenticated     = errors.New("unauthenticated")
	ErrNoMethodSpecified   = errors.New("no method name in context")
	ErrMethodNotFound      = errors.New("method not found")
	ErrNoValidMethodExists = errors.New("no valid method exists")
	ErrUninitialized       = errors.New("uninitialized")
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
	methods     map[string]reflect.Method
}

var _ lambda.Handler = (*Handler)(nil)

var (
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
)

func NewHandler(service interface{}) *Handler {
	return &Handler{
		service: service,

		serviceType: reflect.TypeOf(service),
	}
}

func validMethod(svcType reflect.Type, m reflect.Method) error {
	if m.Type.NumIn() != 3 {
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

	if m.Type.NumOut() != 2 {
		return &ErrInvalidMethod{
			details:  "invalid number of return value",
			errParam: m.Type.NumOut(),
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

func (h *Handler) Init() error {
	methods := make(map[string]reflect.Method)

	for i := 0; i < h.serviceType.NumMethod(); i++ {
		m := h.serviceType.Method(i)
		if err := validMethod(h.serviceType, m); err != nil {
			continue
		}

		methods[m.Name] = m
	}

	if len(methods) == 0 {
		return ErrNoValidMethodExists
	}
	h.methods = methods

	return nil
}

func (h *Handler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	if h.methods == nil {
		return nil, ErrUninitialized
	}

	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return nil, ErrNoLambdaContext
	}

	ctx = context.WithValue(ctx, RequestIdField, lc.AwsRequestID)

	custom := lc.ClientContext.Custom
	if custom == nil {
		return nil, ErrNoClientContext
	}

	apiRequestId, ok := custom[ApiRequestIdField]
	if ok {
		ctx = context.WithValue(ctx, ApiRequestIdField, apiRequestId)
	}

	userId, ok := custom[UserIdField]
	if ok {
		ctx = context.WithValue(ctx, UserIdField, userId)
	}

	functionId, ok := custom[FunctionIdField]
	if !ok {
		return nil, ErrNoMethodSpecified
	}

	m, ok := h.methods[functionId]
	if !ok {
		return nil, ErrMethodNotFound
	}

	reqType := m.Type.In(2)
	if reqType.Kind() == reflect.Ptr {
		reqType = reqType.Elem()
	}

	reqVal := reflect.New(reqType)
	if err := json.Unmarshal(payload, reqVal.Interface()); err != nil {
		return nil, err
	}

	rets := m.Func.Call([]reflect.Value{
		reflect.ValueOf(h.service),
		reflect.ValueOf(ctx),
		reqVal,
	})

	if len(rets) != 2 {
		panic("invalid number of return values")
	}

	resMsg := rets[0].Interface()
	bs, err := json.Marshal(resMsg)
	if err != nil {
		return nil, err
	}

	if e := rets[1].Interface(); e == nil {
		err = nil
	} else {
		err = e.(error)
	}

	return bs, err
}
