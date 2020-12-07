package main

import (
	"context"
	"log"
	"os"

	"github.com/golang/protobuf/proto"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/yunomu/kansousen/lib/awsx"
	db "github.com/yunomu/kansousen/lib/dynamodb"
	"github.com/yunomu/kansousen/lib/lambda"
	apipb "github.com/yunomu/kansousen/proto/api"
)

type server struct {
	table *db.DynamoDB
}

var _ lambda.Server = (*server)(nil)

func (s *server) Serve(ctx context.Context, m proto.Message) (proto.Message, error) {
	name := "unauthorized"
	if v := ctx.Value("username"); v != nil {
		if username, ok := v.(string); ok {
			name = username
		}
	}

	return &apipb.HelloResponse{
		Name:    name,
		Message: "Hello, World!!",
	}, nil
}

func main() {
	ctx := context.Background()

	region := "ap-northeast-1"

	kv, err := awsx.GetSecrets(ctx, region, os.Getenv("SECRET_NAME"))
	if err != nil {
		log.Fatalf("GetSecret: %v", err)
	}

	cred := credentials.NewStaticCredentials(
		kv["AWS_ACCESS_KEY_ID"],
		kv["AWS_SECRET_ACCESS_KEY"],
		"",
	)
	session := session.New(aws.NewConfig().WithCredentials(cred))

	dynamodb := dynamodb.New(session, aws.NewConfig().WithRegion(region))

	table := db.NewDynamoDBTable(dynamodb, "kifulog")

	s := &server{
		table: table,
	}

	h := lambda.NewProtobufHandler((*apipb.HelloRequest)(nil), s)

	h.Start(ctx)
}
