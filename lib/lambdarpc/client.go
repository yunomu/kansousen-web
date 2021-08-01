package lambdarpc

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/pkg/errors"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
)

type Client struct {
	lambdaClient *lambda.Lambda
	functionArn  string

	marshaler      *protojson.MarshalOptions
	unmarshaler    *protojson.UnmarshalOptions
	base64Encoding *base64.Encoding
}

type Option func(*Client)

func SetMarshaler(marshaler *protojson.MarshalOptions) Option {
	return func(c *Client) {
		c.marshaler = marshaler
	}
}

func SetUnmarshaler(unmarshaler *protojson.UnmarshalOptions) Option {
	return func(c *Client) {
		c.unmarshaler = unmarshaler
	}
}

func NewClient(client *lambda.Lambda, functionArn string, opts ...Option) *Client {
	c := &Client{
		lambdaClient: client,
		functionArn:  functionArn,

		marshaler: &protojson.MarshalOptions{
			UseProtoNames: true,
		},
		unmarshaler: &protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
		base64Encoding: base64.URLEncoding,
	}

	for _, f := range opts {
		f(c)
	}

	return c
}

type LambdaError struct {
	ErrorType    string `json:"errorType"`
	ErrorMessage string `json:"errorMessage"`
}

func (e *LambdaError) Error() string {
	return e.ErrorType + ": " + e.ErrorMessage
}

func (c *Client) Invoke(ctx context.Context, requestId string, in, out proto.Message) error {
	bs, err := c.marshaler.Marshal(in)
	if err != nil {
		return errors.Wrap(err, "json.Marshal(in)")
	}

	o, err := c.lambdaClient.InvokeWithContext(ctx, &lambda.InvokeInput{
		ClientContext:  aws.String(c.base64Encoding.EncodeToString([]byte(requestId))),
		FunctionName:   aws.String(c.functionArn),
		InvocationType: aws.String(lambda.InvocationTypeRequestResponse),
		Payload:        bs,
	})
	if err != nil {
		return errors.Wrap(err, "lambda.Invoke")
	}

	if o.FunctionError != nil {
		buf := bytes.NewBuffer(o.Payload)
		d := json.NewDecoder(buf)
		errObj := &LambdaError{}
		if err := d.Decode(errObj); err != nil {
			return errors.Wrapf(err, "Error payload decoder error: `%s`", buf.String())
		}

		return errObj
	}

	if err := c.unmarshaler.Unmarshal(o.Payload, out); err != nil {
		return errors.Wrap(err, "json.Unmarshal(out)")
	}

	return nil
}
