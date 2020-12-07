package dynamodb

type Logger interface {
	Info(function string, message string)
}

type nopLogger struct{}

var _ Logger = (*nopLogger)(nil)

func (l *nopLogger) Info(string, string) {}
