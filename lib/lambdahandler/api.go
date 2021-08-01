package lambdahandler

import (
	"context"
	"encoding/json"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type RequestContext struct {
	RequestId string
	UserId    string
}

type Request events.APIGatewayProxyRequest
type Response events.APIGatewayProxyResponse

type Error interface {
	statusCode() int
	errorType() string
	errorMessage() string
	Error() string
}

type serverError struct{}

func ServerError() Error {
	return &serverError{}
}

var _ Error = (*serverError)(nil)

func (*serverError) statusCode() int {
	return 500
}

func (*serverError) errorType() string {
	return "ServerError"
}

func (*serverError) errorMessage() string {
	return "Internal Server Error"
}

func (e *serverError) Error() string {
	return e.errorMessage()
}

type clientError struct {
	code int
	msg  string
}

func ClientError(statusCode int, errorMessage string) Error {
	return &clientError{
		code: statusCode,
		msg:  errorMessage,
	}
}

var _ Error = (*clientError)(nil)

func (c *clientError) statusCode() int {
	return c.code
}

func (*clientError) errorType() string {
	return "ClientError"
}

func (c *clientError) errorMessage() string {
	return c.msg
}

func (e *clientError) Error() string {
	return e.errorMessage()
}

type handler func(context.Context, *RequestContext, *Request) (proto.Message, Error)

type Logger interface {
	Error(msg string, err error)
}

type defaultLogger struct{}

var _ Logger = (*defaultLogger)(nil)

func (*defaultLogger) Error(msg string, err error) {}

type APIHandler struct {
	marshaler *protojson.MarshalOptions
	handlers  map[string]map[string]handler
	logger    Logger
}

type APIHandlerOption func(*APIHandler)

func SetMarshaler(marshaler *protojson.MarshalOptions) APIHandlerOption {
	return func(s *APIHandler) {
		s.marshaler = marshaler
	}
}

func AddAPIHandler(path, method string, h func(context.Context, *RequestContext, *Request) (proto.Message, Error)) APIHandlerOption {
	return func(s *APIHandler) {
		p, ok := s.handlers[path]
		if !ok {
			p = map[string]handler{}
		}

		p[method] = h
		s.handlers[path] = p
	}
}

func SetLogger(logger Logger) APIHandlerOption {
	return func(s *APIHandler) {
		if logger == nil {
			logger = &defaultLogger{}
		}
		s.logger = logger
	}
}

func NewAPIHandler(opts ...APIHandlerOption) *APIHandler {
	h := &APIHandler{
		marshaler: &protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
		handlers: map[string]map[string]handler{},
		logger:   &defaultLogger{},
	}

	for _, f := range opts {
		f(h)
	}

	return h
}

func (s *APIHandler) buildResponse(statusCode int, contentType, body string) *Response {
	headers := map[string]string{
		"Access-Control-Allow-Origin": "*",
	}

	if contentType != "" {
		headers["Content-Type"] = contentType
	}

	return &Response{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body,
	}
}

type ErrorMessage struct {
	ErrorType    string `json:"error_type"`
	ErrorMessage string `json:"error_message"`
}

type ErrorDecodeError struct {
	OriginalErrorMessage *ErrorMessage `json:"original_error_message"`
	Err                  error         `json:"err"`
}

func (e *ErrorDecodeError) Error() string {
	if e.Err == nil {
		return "ErrorDecodeError"
	}
	return e.Err.Error()
}

func (s *APIHandler) errorResponse(e Error) *Response {
	msg := &ErrorMessage{
		ErrorType:    e.errorType(),
		ErrorMessage: e.errorMessage(),
	}

	var buf strings.Builder
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(msg); err != nil {
		s.logger.Error("Error decode", &ErrorDecodeError{
			OriginalErrorMessage: msg,
			Err:                  err,
		})
		return s.buildResponse(500, "application/json", `{"error_message":"Internal Server Error"}`)
	}

	return s.buildResponse(e.statusCode(), "application/josn", buf.String())
}

func (s *APIHandler) response(msg proto.Message) *Response {
	var body, contentType string
	if msg == nil {
		bs, err := s.marshaler.Marshal(msg)
		if err != nil {
			s.logger.Error("protojson.Marshal(response)", err)
			return s.errorResponse(ServerError())
		}

		body = string(bs)
		contentType = "application/json"
	}

	return s.buildResponse(200, contentType, body)
}

type AuthorizerError struct {
	Message        string                                `json:"message"`
	RequestContext *events.APIGatewayProxyRequestContext `json:"request_context"`
}

func (e *AuthorizerError) Error() string {
	return e.Message
}

func getRequestContext(ctx *events.APIGatewayProxyRequestContext) (*RequestContext, error) {
	requestId := ctx.RequestID

	claimsVal, ok := ctx.Authorizer["claims"]
	if !ok {
		return nil, &AuthorizerError{
			Message:        "claims not found",
			RequestContext: ctx,
		}
	}

	claims, ok := claimsVal.(map[string]interface{})
	if !ok {
		return nil, &AuthorizerError{
			Message:        "unknown claims format",
			RequestContext: ctx,
		}
	}

	userIdVal, ok := claims["sub"]
	if !ok {
		return nil, &AuthorizerError{
			Message:        "claims sub not found",
			RequestContext: ctx,
		}
	}

	userId, ok := userIdVal.(string)
	if !ok {
		return nil, &AuthorizerError{
			Message:        "unknown claims sub format",
			RequestContext: ctx,
		}
	}

	return &RequestContext{
		RequestId: requestId,
		UserId:    userId,
	}, nil
}

func (s *APIHandler) handle(ctx context.Context, req *Request) (*Response, error) {
	reqCtx, err := getRequestContext(&req.RequestContext)
	if err != nil {
		s.logger.Error("sub is not found in claims", err)
		return s.errorResponse(ServerError()), nil
	}

	path, ok := s.handlers[req.Path]
	if !ok {
		return s.errorResponse(ClientError(404, "NotFound")), nil
	}

	h, ok := path[req.HTTPMethod]
	if !ok {
		return s.errorResponse(ClientError(405, "MethodNotAllowed")), nil
	}

	msg, err := h(ctx, reqCtx, req)
	if err != nil {
		return nil, err
	}

	return s.response(msg), nil
}

func (s *APIHandler) Start(ctx context.Context) {
	lambda.StartWithContext(ctx, s.handle)
}
