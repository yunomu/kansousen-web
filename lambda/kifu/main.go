package main

import (
	"context"
	"os"

	"go.uber.org/zap"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsdynamodb "github.com/aws/aws-sdk-go/service/dynamodb"

	libdb "github.com/yunomu/kansousen/lib/db"
	libdynamodb "github.com/yunomu/kansousen/lib/dynamodb"
	"github.com/yunomu/kansousen/service/kifu"

	"github.com/yunomu/kansousen/proto/lambdakifu"
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

type handler struct {
	service     *kifu.Service
	unmarshaler *protojson.UnmarshalOptions
	marshaler   *protojson.MarshalOptions
}

func (h *handler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	in := &lambdakifu.Input{}
	if err := h.unmarshaler.Unmarshal(payload, in); err != nil {
		return nil, err
	}

	output := &lambdakifu.Output{}
	switch t := in.Select.(type) {
	case *lambdakifu.Input_GetKifuInput:
		out, err := h.getKifu(ctx, t.GetKifuInput)
		if err != nil {
			return nil, err
		}
		output.Select = &lambdakifu.Output_GetKifuOutput{
			GetKifuOutput: out,
		}
	case *lambdakifu.Input_PostKifuInput:
		out, err := h.postKifu(ctx, t.PostKifuInput)
		if err != nil {
			return nil, err
		}
		output.Select = &lambdakifu.Output_PostKifuOutput{
			PostKifuOutput: out,
		}
	case *lambdakifu.Input_RecentKifuInput:
		out, err := h.recentKifu(ctx, t.RecentKifuInput)
		if err != nil {
			return nil, err
		}
		output.Select = &lambdakifu.Output_RecentKifuOutput{
			RecentKifuOutput: out,
		}
	case *lambdakifu.Input_DeleteKifuInput:
		if err := h.deleteKifu(ctx, t.DeleteKifuInput); err != nil {
			return nil, err
		}
		output.Select = &lambdakifu.Output_DeleteKifuOutput{}
	}

	return h.marshaler.Marshal(output)
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

	dynamodb := awsdynamodb.New(session, aws.NewConfig().WithRegion(region))

	table := libdynamodb.NewDynamoDBTable(dynamodb, tableName)
	if err := table.Init(ctx); err != nil {
		zap.L().Fatal("DynamoDBTable.Init", zap.Error(err), zap.String("tableName", tableName))
	}

	h := &handler{
		service:   kifu.NewService(libdb.NewDynamoDB(table)),
		marshaler: &protojson.MarshalOptions{},
		unmarshaler: &protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}

	lambda.StartHandlerWithContext(ctx, h)
}
