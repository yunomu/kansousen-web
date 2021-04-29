package lambda

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/yunomu/kansousen/lib/jwt"
)

type Request events.APIGatewayProxyRequest
type Response events.APIGatewayProxyResponse

type Server interface {
	Serve(context.Context, []byte) ([]byte, error)
}

type Handler struct {
	server Server
}

func NewHandler(s Server) *Handler {
	return &Handler{
		server: s,
	}
}

func (h *Handler) handler(ctx context.Context, r Request) (Response, error) {
	header := map[string]string{
		"Access-Control-Allow-Origin": "*",
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
			err := errors.New("unauthorized")
			return Response{
				StatusCode: http.StatusInternalServerError,
				Headers:    header,
				Body:       err.Error(),
			}, err
		}

		ctx = context.WithValue(ctx, "userId", username)
	}

	res, err := h.server.Serve(ctx, []byte(r.Body))
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

	header["Content-Type"] = "application/json"
	return Response{
		StatusCode: http.StatusOK,
		Headers:    header,
		Body:       string(res),
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

func (h *Handler) Start(ctx context.Context) {
	lambda.StartWithContext(ctx, h.handler)
}
