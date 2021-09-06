package main

import (
	"context"
	"os"

	"go.uber.org/zap"

	lambdaclient "github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"

	"github.com/yunomu/kansousen/lib/config"
	"github.com/yunomu/kansousen/lib/lambda/lambdagateway"
	"github.com/yunomu/kansousen/lib/lambda/lambdarpc"
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

type apiLogger struct{}

func (*apiLogger) Error(msg string, err error) {
	zap.L().Error(msg, zap.Error(err))
}

func main() {
	ctx := context.Background()

	configURL := os.Getenv("CONFIG_URL")
	cfg, err := config.Load(configURL)
	if err != nil {
		zap.L().Fatal("Load config error", zap.Error(err), zap.String("configURL", configURL))
	}

	session := session.New()
	lambdaClient := lambda.New(session, aws.NewConfig().WithRegion(cfg["Region"]))

	gw := lambdagateway.NewLambdaGateway(lambdaClient,
		lambdagateway.WithAPIRequestID(lambdarpc.ApiRequestIdField),
		lambdagateway.WithClaimSubID(lambdarpc.UserIdField),
		lambdagateway.AddFunction("/v1/post-kifu", "POST", cfg["KifuFunction"], "PostKifu"),
		lambdagateway.AddFunction("/v1/get-kifu", "POST", cfg["KifuFunction"], "GetKifu"),
		lambdagateway.AddFunction("/v1/delete-kifu", "POST", cfg["KifuFunction"], "DeleteKifu"),
		lambdagateway.AddFunction("/v1/recent-kifu", "POST", cfg["KifuFunction"], "RecentKifu"),
		lambdagateway.AddFunction("/v1/same-positions", "POST", cfg["KifuFunction"], "GetSamePositions"),
		lambdagateway.SetLogger(&apiLogger{}),
		lambdagateway.SetFunctionErrorHandler(func(e *lambdagateway.LambdaError) error {
			switch e.ErrorType {
			case "InvalidArgumentError":
				return lambdagateway.ClientError(400, e.ErrorMessage)
			default:
				zap.L().Error("lambda.Invoke", zap.Any("error", e))
				return lambdagateway.ServerError()
			}
		}),
	)

	lambdaclient.StartWithContext(ctx, gw.Serve)
}
