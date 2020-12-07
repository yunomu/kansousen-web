package lambda

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/yunomu/kansousen/lib/jwt"
)

type Request events.APIGatewayProxyRequest
type Response events.APIGatewayProxyResponse

type Server interface {
	Serve(context.Context, proto.Message) (proto.Message, error)
}

type ProtobufHandler struct {
	marshaler   *jsonpb.Marshaler
	unmarshaler *jsonpb.Unmarshaler

	reqType reflect.Type
	server  Server
}

func NewProtobufHandler(requestTemplate proto.Message, s Server) *ProtobufHandler {
	return &ProtobufHandler{
		marshaler: &jsonpb.Marshaler{
			EmitDefaults: true,
		},
		unmarshaler: &jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		},
		reqType: reflect.TypeOf(requestTemplate).Elem(),
		server:  s,
	}
}

func (h *ProtobufHandler) handler(ctx context.Context, r Request) (Response, error) {
	header := map[string]string{
		"Access-Control-Allow-Origin": "*",
	}

	i := reflect.New(h.reqType).Interface()

	req, ok := i.(proto.Message)
	if !ok {
		return Response{
			StatusCode: http.StatusServiceUnavailable,
			Headers:    header,
			Body:       "failed precondition",
		}, errors.New("request is not instance of proto.Message")
	}

	if err := h.unmarshaler.Unmarshal(strings.NewReader(r.Body), req); err != nil {
		return Response{
			StatusCode: http.StatusBadRequest,
			Headers:    header,
			Body:       err.Error(),
		}, err
	}

	if auth, ok := r.Headers["Authorization"]; ok {
		payload, err := jwt.DecodePayload(auth)
		if err != nil {
			return Response{
				StatusCode: http.StatusInternalServerError,
				Headers:    header,
				Body:       err.Error(),
			}, err
		}

		username, ok := payload["cognito:username"]
		if !ok {
			err := errors.New("username is not found")
			return Response{
				StatusCode: http.StatusInternalServerError,
				Headers:    header,
				Body:       err.Error(),
			}, err
		}

		ctx = context.WithValue(ctx, "userId", username)
	}

	res, err := h.server.Serve(ctx, req)
	switch code := status.Code(err); code {
	case codes.OK:
		// do nothing
	case codes.Unknown:
		return Response{
			StatusCode: http.StatusInternalServerError,
			Headers:    header,
		}, err
	default:
		return Response{
			StatusCode: runtime.HTTPStatusFromCode(code),
			Headers:    header,
			Body:       status.Convert(err).Message(),
		}, err
	}

	body, err := h.marshaler.MarshalToString(res)
	if err != nil {
		return Response{
			StatusCode: http.StatusInternalServerError,
			Headers:    header,
		}, err
	}

	header["Content-Type"] = "application/json"
	return Response{
		StatusCode: http.StatusOK,
		Headers:    header,
		Body:       body,
	}, nil
}

func GetUserId(ctx context.Context) string {
	if v := ctx.Value("userId"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}

	return ""
}

func (h *ProtobufHandler) Start(ctx context.Context) {
	lambda.StartWithContext(ctx, h.handler)
}
