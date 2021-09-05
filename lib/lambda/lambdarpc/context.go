package lambdarpc

import (
	"context"
)

const (
	ApiRequestIdField = "api-request-id"
	RequestIdField    = "request-id"
	UserIdField       = "user-id"
	FunctionIdField   = "function-id"
)

func GetApiRequestId(ctx context.Context) string {
	v := ctx.Value(ApiRequestIdField)
	return v.(string)
}

func GetRequestId(ctx context.Context) string {
	v := ctx.Value(RequestIdField)
	return v.(string)
}

func GetUserId(ctx context.Context) string {
	v := ctx.Value(UserIdField)
	return v.(string)
}
