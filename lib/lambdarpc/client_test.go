package lambdarpc

import (
	"testing"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

func TestClient_encodeClientContext(t *testing.T) {
	cc := &lambdacontext.ClientContext{
		Custom: map[string]string{
			"key": "test",
		},
	}

	client := NewClient(nil, "")
	str, err := client.encodeClientContext(cc)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	exp := "eyJDbGllbnQiOnsiaW5zdGFsbGF0aW9uX2lkIjoiIiwiYXBwX3RpdGxlIjoiIiwiYXBwX3ZlcnNpb25fY29kZSI6IiIsImFwcF9wYWNrYWdlX25hbWUiOiIifSwiZW52IjpudWxsLCJjdXN0b20iOnsia2V5IjoidGVzdCJ9fQo="
	if str != exp {
		t.Errorf("mismatch: exp=%v act=%v", exp, str)
	}
}
