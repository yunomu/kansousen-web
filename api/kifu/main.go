package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"

	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"

	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"

	apipb "github.com/yunomu/kansousen/proto/api"

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

type server struct {
	kifuFuncArn  string
	lambdaClient *lambda.Lambda

	marshaler   *protojson.MarshalOptions
	unmarshaler *protojson.UnmarshalOptions
}

func convRequest(userId string, req *apipb.KifuRequest) (*lambdakifu.Input, *response) {
	in := &lambdakifu.Input{}

	switch t := req.KifuRequestSelect.(type) {
	case *apipb.KifuRequest_RequestRecentKifu:
		r := t.RequestRecentKifu
		in.Select = &lambdakifu.Input_RecentKifuInput{
			RecentKifuInput: &lambdakifu.RecentKifuInput{
				UserId: userId,
				Limit:  r.Limit,
			},
		}
	case *apipb.KifuRequest_RequestPostKifu:
		r := t.RequestPostKifu

		encoding, ok := lambdakifu.PostKifuInput_Encoding_value[r.Encoding]
		if !ok {
			return nil, clientError(400, "unknown encoding")
		}

		format, ok := lambdakifu.PostKifuInput_Format_value[r.Format]
		if !ok {
			return nil, clientError(400, "unknown format")
		}

		in.Select = &lambdakifu.Input_PostKifuInput{
			PostKifuInput: &lambdakifu.PostKifuInput{
				UserId:   userId,
				Encoding: lambdakifu.PostKifuInput_Encoding(encoding),
				Format:   lambdakifu.PostKifuInput_Format(format),
				Payload:  r.Payload,
			},
		}
	case *apipb.KifuRequest_RequestDeleteKifu:
		r := t.RequestDeleteKifu
		in.Select = &lambdakifu.Input_DeleteKifuInput{
			DeleteKifuInput: &lambdakifu.DeleteKifuInput{
				UserId: userId,
				KifuId: r.KifuId,
			},
		}
	case *apipb.KifuRequest_RequestGetKifu:
		r := t.RequestGetKifu
		in.Select = &lambdakifu.Input_GetKifuInput{
			GetKifuInput: &lambdakifu.GetKifuInput{
				UserId: userId,
				KifuId: r.KifuId,
			},
		}
	case *apipb.KifuRequest_RequestGetSamePositions:
		r := t.RequestGetSamePositions
		in.Select = &lambdakifu.Input_GetSamePositionsInput{
			GetSamePositionsInput: &lambdakifu.GetSamePositionsInput{
				UserId:         userId,
				Position:       r.Position,
				Steps:          r.Steps,
				ExcludeKifuIds: r.ExcludeKifuIds,
			},
		}
	default:
		return nil, clientError(400, "unknown request")
	}

	return in, nil
}

func convResponse(out *lambdakifu.Output) *apipb.KifuResponse {
	res := &apipb.KifuResponse{}

	switch t := out.Select.(type) {
	case *lambdakifu.Output_GetKifuOutput:
		o := t.GetKifuOutput
		var _ = o
	case *lambdakifu.Output_RecentKifuOutput:
		o := t.RecentKifuOutput
		var kifus []*apipb.RecentKifuResponse_Kifu
		for _, kifu := range o.Kifus {
			kifus = append(kifus, &apipb.RecentKifuResponse_Kifu{
				UserId:  kifu.UserId,
				KifuId:  kifu.KifuId,
				StartTs: kifu.StartTs,

				Handicap:     kifu.Handicap,
				GameName:     kifu.GameName,
				FirstPlayer:  strings.Join(kifu.FirstPlayers, ", "),
				SecondPlayer: strings.Join(kifu.SecondPlayers, ", "),
				Note:         kifu.Note,
			})
		}

		res.KifuResponseSelect = &apipb.KifuResponse_ResponseRecentKifu{
			ResponseRecentKifu: &apipb.RecentKifuResponse{
				Kifus: kifus,
			},
		}
	case *lambdakifu.Output_PostKifuOutput:
		o := t.PostKifuOutput
		res.KifuResponseSelect = &apipb.KifuResponse_ResponsePostKifu{
			ResponsePostKifu: &apipb.PostKifuResponse{
				KifuId: o.KifuId,
			},
		}
	case *lambdakifu.Output_DeleteKifuOutput:
		res.KifuResponseSelect = &apipb.KifuResponse_ResponseDeleteKifu{
			ResponseDeleteKifu: &apipb.DeleteKifuResponse{},
		}
	case *lambdakifu.Output_GetSamePositionsOutput:
	default:
		zap.L().Error("convResponse: unknown type", zap.Any("type", t))
		return nil
	}

	return res
}

type request events.APIGatewayProxyRequest
type response events.APIGatewayProxyResponse

func getUserId(ctx *events.APIGatewayProxyRequestContext) string {
	claimsVal, ok := ctx.Authorizer["claims"]
	if !ok {
		return ""
	}

	claims, ok := claimsVal.(map[string]interface{})
	if !ok {
		return ""
	}

	userIdVal, ok := claims["sub"]
	if !ok {
		return ""
	}

	userId, ok := userIdVal.(string)
	if !ok {
		return ""
	}

	return userId
}

func buildResponse(statusCode int, contentType string, body string) *response {
	headers := map[string]string{
		"Access-Control-Allow-Origin": "*",
	}

	if contentType != "" {
		headers["Content-Type"] = contentType
	}

	return &response{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body,
	}
}

func clientError(statusCode int, msg string) *response {
	return buildResponse(statusCode, "", msg)
}

func serverError() *response {
	return buildResponse(500, "", "\"Internal Server Error\"")
}

func (s *server) kifu(ctx context.Context, r *request) *response {
	userId := getUserId(&r.RequestContext)
	if userId == "" {
		zap.L().Error("sub is not found in claims", zap.Any("RequestContext", &r.RequestContext))
		return serverError()
	}

	req := &apipb.KifuRequest{}
	if err := s.unmarshaler.Unmarshal([]byte(r.Body), req); err != nil {
		return clientError(400, err.Error())
	}

	in, errRes := convRequest(userId, req)
	if errRes != nil {
		return errRes
	}

	bs, err := s.marshaler.Marshal(in)
	if err != nil {
		zap.L().Error("json.Marshal(in)", zap.Error(err))
		return serverError()
	}

	o, err := s.lambdaClient.InvokeWithContext(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(s.kifuFuncArn),
		InvocationType: aws.String(lambda.InvocationTypeRequestResponse),
		Payload:        bs,
	})
	if err != nil {
		zap.L().Error("lambda.Invoke", zap.Error(err))
		return serverError()
	}

	if o.FunctionError != nil {
		d := json.NewDecoder(bytes.NewReader(o.Payload))
		errObj := map[string]interface{}{}
		if err := d.Decode(&errObj); err != nil {
			zap.L().Error("json.Decode", zap.Error(err), zap.ByteString("Payload", o.Payload))
			return serverError()
		}

		// TODO: v["errorType"]判別
		var fields []zap.Field
		for k, v := range errObj {
			fields = append(fields, zap.Any("payload_"+k, v))
		}
		zap.L().Error("FunctionError", fields...)
		return serverError()
	}

	out := &lambdakifu.Output{}
	if err := s.unmarshaler.Unmarshal(o.Payload, out); err != nil {
		zap.L().Error("json.Unmarshal(out)", zap.Error(err))
		return serverError()
	}

	res := convResponse(out)
	if res == nil {
		return serverError()
	}

	outBs, err := s.marshaler.Marshal(res)
	if err != nil {
		zap.L().Error("json.Marshal(response)", zap.Error(err))
		return serverError()
	}

	return buildResponse(200, "application/json", string(outBs))
}

func (s *server) handler(ctx context.Context, req *request) (*response, error) {
	switch req.HTTPMethod {
	case "POST":
		switch req.Path {
		case "/v1/kifu":
			return s.kifu(ctx, req), nil
		default:
			return clientError(404, ""), nil
		}
	default:
		return clientError(405, ""), nil
	}
}

func main() {
	ctx := context.Background()

	kifuFuncArn := os.Getenv("KIFU_FUNC_ARN")
	if kifuFuncArn == "" {
		zap.L().Fatal("env KIFU_FUNC_ARN is not found")
	}

	region := os.Getenv("REGION")

	session := session.New()
	lambdaClient := lambda.New(session, aws.NewConfig().WithRegion(region))

	s := &server{
		kifuFuncArn:  kifuFuncArn,
		lambdaClient: lambdaClient,
		marshaler:    &protojson.MarshalOptions{},
		unmarshaler: &protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}

	runtime.StartWithContext(ctx, s.handler)
}
