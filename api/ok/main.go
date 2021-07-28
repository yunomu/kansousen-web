package main

import (
	"context"
	"encoding/json"
	"os"

	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

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
	region    string
	tableName string
	db        *dynamodb.DynamoDB
}

type request events.APIGatewayProxyRequest
type response events.APIGatewayProxyResponse

func retval(body map[string]interface{}) (*response, error) {
	bs, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return &response{
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
			"Content-Type":                "application/json",
		},
		Body: string(bs),
	}, nil
}

func (s *server) healthz(ctx context.Context, req *request) (*response, error) {

	ret := map[string]interface{}{}

	//session := session.New()
	//db := dynamodb.New(session, aws.NewConfig().WithRegion(s.region))

	out, err := s.db.DescribeTableWithContext(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(s.tableName),
	})
	if err != nil {
		ret["status"] = "NG"
		return retval(ret)
	}

	ret["status"] = "OK"
	ret["table_status"] = out.Table.TableStatus
	return retval(ret)
}

func (s *server) handler(ctx context.Context, req *request) (*response, error) {
	header := map[string]string{
		"Access-Control-Allow-Origin": "*",
	}

	switch req.HTTPMethod {
	case "GET":
		switch req.Path {
		case "/v1/ok":
			return s.healthz(ctx, req)
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

	region := os.Getenv("REGION")
	tableName := os.Getenv("TABLE_NAME")

	session := session.New()
	db := dynamodb.New(session, aws.NewConfig().WithRegion(region))

	s := &server{
		region:    region,
		tableName: tableName,
		db:        db,
	}

	runtime.StartWithContext(ctx, s.handler)
}
