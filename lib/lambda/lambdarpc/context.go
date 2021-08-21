package lambdarpc

type Context struct {
	ApiRequestId string
	RequestId    string
	UserId       string
}

const (
	ApiRequestIdField = "api-request-id"
	UserIdField       = "user-id"
	MethodField       = "method"
)

func FromCustomMap(custom map[string]string) *Context {
	if custom == nil {
		return &Context{}
	}

	return &Context{
		ApiRequestId: custom[ApiRequestIdField],
		UserId:       custom[UserIdField],
	}
}
