package lambdarpc

type lambdarpcError interface {
	error
	lambdarpcError()
}

type ClientError struct {
	Message string
	Err     error
}

func (e *ClientError) Error() string {
	switch {
	case e.Err == nil:
		return e.Message
	case e.Message == "":
		return e.Err.Error()
	default:
		return e.Message + ": " + e.Err.Error()
	}
}

func (*ClientError) lambdarpcError() {}

type InternalError struct {
	Message string
	Err     error
}

func (e *InternalError) Error() string {
	switch {
	case e.Err == nil:
		return e.Message
	case e.Message == "":
		return e.Err.Error()
	default:
		return e.Message + ": " + e.Err.Error()
	}
}

func (*InternalError) lambdarpcError() {}
