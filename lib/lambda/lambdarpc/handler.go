package lambdarpc

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

type InvalidMethodError struct {
	Details      string
	InvalidParam interface{}
}

func (e *InvalidMethodError) Error() string {
	return "invalid method signature: func (obj) (context.Context, *Context, proto.Message) (proto.Message, errors)"
}

type Handler struct {
	service interface{}

	serviceType reflect.Type
	methods     map[string]reflect.Method
	initialized bool
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
		return &InvalidMethodError{
			Details:      "invalid parameter number",
			InvalidParam: m.Type.NumIn(),
		}
	}

	if m.Type.In(0).String() != svcType.String() {
		return &InvalidMethodError{
			Details:      "invalid service type",
			InvalidParam: m.Type.In(0),
		}
	}

	if m.Type.In(1).String() != contextType.String() {
		return &InvalidMethodError{
			Details:      "1st param is not context.Context",
			InvalidParam: m.Type.In(1),
		}
	}

	if m.Type.NumOut() != 2 {
		return &InvalidMethodError{
			Details:      "invalid number of return value",
			InvalidParam: m.Type.NumOut(),
		}
	}

	if m.Type.Out(1).String() != errorType.String() {
		return &InvalidMethodError{
			Details:      "2nd return value is not error",
			InvalidParam: m.Type.Out(1),
		}
	}

	return nil
}

func (h *Handler) Init() error {
	if h.initialized {
		return nil
	}

	methods := make(map[string]reflect.Method)

	for i := 0; i < h.serviceType.NumMethod(); i++ {
		m := h.serviceType.Method(i)
		if err := validMethod(h.serviceType, m); err != nil {
			continue
		}

		methods[m.Name] = m
	}

	if len(methods) == 0 {
		return &InvalidMethodError{
			Details: "no valid method exists",
		}
	}
	h.methods = methods

	h.initialized = true

	return nil
}

func (h *Handler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	if err := h.Init(); err != nil {
		return nil, err
	}

	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return nil, &InternalError{
			Message: "No lambda context",
		}
	}

	ctx = context.WithValue(ctx, RequestIdField, lc.AwsRequestID)

	custom := lc.ClientContext.Custom
	if custom == nil {
		return nil, &InternalError{
			Message: "No client context",
		}
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
		return nil, &ClientError{
			Message: "No method specified",
		}
	}

	m, ok := h.methods[functionId]
	if !ok {
		return nil, &ClientError{
			Message: "method not found",
		}
	}

	reqType := m.Type.In(2)
	if reqType.Kind() == reflect.Ptr {
		reqType = reqType.Elem()
	}

	reqVal := reflect.New(reqType)
	if err := json.Unmarshal(payload, reqVal.Interface()); err != nil {
		return nil, &InternalError{
			Message: "json.Unmarshal(req)",
			Err:     err,
		}
	}

	rets := m.Func.Call([]reflect.Value{
		reflect.ValueOf(h.service),
		reflect.ValueOf(ctx),
		reqVal,
	})

	if len(rets) != 2 {
		return nil, &InternalError{
			Message: "invalid number of return values",
		}
	}

	resMsg := rets[0].Interface()
	bs, err := json.Marshal(resMsg)
	if err != nil {
		return nil, &InternalError{
			Message: "json.Marshal(res)",
			Err:     err,
		}
	}

	if e := rets[1].Interface(); e != nil {
		if err, ok := e.(lambdarpcError); ok {
			return nil, err
		}
		return nil, &InternalError{
			Err: e.(error),
		}
	}

	return bs, nil
}
