package main

import (
	"context"
	"os"

	"go.uber.org/zap"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/yunomu/kansousen/lib/config"
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

	configURL := os.Getenv("CONFIG_URL")
	cfg, err := config.Load(configURL)
	if err != nil {
		zap.L().Fatal("Load config error", zap.Error(err), zap.String("configURL", configURL))
	}

	session := session.New()

	zap.L().Info("Start",
		zap.String("region", cfg["Region"]),
		zap.String("table_name", cfg["KifuTable"]),
	)

	dynamodb := dynamodb.New(session, aws.NewConfig().WithRegion(cfg["Region"]))
	table := db.NewDynamoDB(dynamodb, cfg["KifuTable"])
	svc := service.NewService(table)

	h := lambdarpc.NewHandler(svc)

	lambda.StartHandlerWithContext(ctx, h)
}
