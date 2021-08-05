package requestcontext

type Context struct {
	RequestId string
	UserId    string
}

const (
	requestIdField = "request-context-request-id"
	userIdField    = "request-context-user-id"
)

func (c *Context) Encode(out map[string]string) {
	if c == nil || out == nil {
		return
	}

	out[requestIdField] = c.RequestId
	out[userIdField] = c.UserId
}

func Decode(in map[string]string) *Context {
	if in == nil {
		return nil
	}

	return &Context{
		RequestId: in[requestIdField],
		UserId:    in[userIdField],
	}
}
