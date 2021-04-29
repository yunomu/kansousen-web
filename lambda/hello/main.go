package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/aws/aws-lambda-go/lambda"
)

type handler struct {
	db *dynamodb.DynamoDB
}

func (h *handler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	res, err := h.db.ListTablesWithContext(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return nil, err
	}

	var ret []string
	for _, t := range res.TableNames {
		ret = append(ret, aws.StringValue(t))
	}

	return json.Marshal(ret)
}

func main() {
	ctx := context.Background()

	sess := session.New()

	dynamodb := dynamodb.New(sess)

	h := &handler{
		db: dynamodb,
	}
	lambda.StartHandlerWithContext(ctx, h)
}
