package main

import (
	"context"
	"os"

	"go.uber.org/zap"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/yunomu/kansousen/lib/db"
	"github.com/yunomu/kansousen/lib/lambda/lambdarpc"

	"github.com/yunomu/kansousen/lambda/kifu/service"
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

func main() {
	ctx := context.Background()

	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		zap.L().Fatal("env TABLE_NAME is not found")
	}

	region := os.Getenv("REGION")

	session := session.New()

	zap.L().Info("Start",
		zap.String("region", region),
		zap.String("table_name", tableName),
	)

	dynamodb := dynamodb.New(session, aws.NewConfig().WithRegion(region))
	table := db.NewDynamoDB(dynamodb, tableName)
	svc := service.NewService(table)

	h := lambdarpc.NewHandler(svc)

	lambda.StartHandlerWithContext(ctx, h)
}
