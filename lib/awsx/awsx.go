package awsx

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

func GetSecrets(ctx context.Context, region, secretName string) (map[string]string, error) {
	sm := secretsmanager.New(session.New(), aws.NewConfig().WithRegion(region))
	out, err := sm.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	})
	if err != nil {
		return nil, err
	}

	var in io.Reader
	if out.SecretString != nil {
		in = bytes.NewBufferString(aws.StringValue(out.SecretString))
	} else {
		in = bytes.NewBuffer(out.SecretBinary)
		in = base64.NewDecoder(base64.StdEncoding, in)
	}

	kv := make(map[string]string)
	if err := json.NewDecoder(in).Decode(&kv); err != nil {
		return nil, err
	}

	return kv, nil
}
