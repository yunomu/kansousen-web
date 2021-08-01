package lambdarpc

import (
	"testing"

	"encoding/base64"
)

func TestClientContext_Encode(t *testing.T) {
	ctx := &clientContext{
		RequestId: "test",
	}

	str, err := ctx.encode(base64.URLEncoding)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	exp := "eyJyZXF1ZXN0X2lkIjoidGVzdCJ9"
	if str != exp {
		t.Errorf("mismatch: exp=%v act=%v", exp, str)
	}
}
