package db

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func TestAttributeValue(t *testing.T) {
	av, err := dynamodbattribute.MarshalMap(DynamoDBKifuRecord{
		KifuId: "test-kifu-id",
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if len(av) != 2 || av["kifuId"] == nil || av["seq"] == nil {
		t.Fatalf("av: %#v", av)
	}
}
