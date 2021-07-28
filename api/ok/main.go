package main

import (
	"context"
	"os"

	"go.uber.org/zap"

	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
)

func init() {
	var logger *zap.Logger
	if os.Getenv("DEV") == "true" {
		l, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		logger = l
	} else {
		l, err := zap.NewProduction()
		if err != nil {
			panic(err)
		}
		logger = l
	}
	zap.ReplaceGlobals(logger)
}

type server struct {
}

type request events.APIGatewayProxyRequest
type response events.APIGatewayProxyResponse

func (s *server) handler(ctx context.Context, req *request) (*response, error) {
	header := map[string]string{
		"Access-Control-Allow-Origin": "*",
	}

	switch req.HTTPMethod {
	case "GET":
		switch req.Path {
		case "/v1/ok":
			header["Content-Type"] = "application/json"
			return &response{
				StatusCode: 200,
				Headers:    header,
				Body:       `{"status":"OK"}`,
			}, nil
		default:
			return &response{
				StatusCode: 404,
				Headers:    header,
			}, nil
		}
	default:
		return &response{
			StatusCode: 405,
			Headers:    header,
		}, nil
	}
}

func main() {
	ctx := context.Background()

	s := &server{}

	runtime.StartWithContext(ctx, s.handler)
}
