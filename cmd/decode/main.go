package decode

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/google/subcommands"
)

type Command struct {
}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Name() string     { return "decode" }
func (c *Command) Synopsis() string { return "decode authentication result" }
func (c *Command) Usage() string {
	return `
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	f.SetOutput(os.Stderr)
}

func fieldDecode(s string) (map[string]interface{}, error) {
	bs, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewReader(bs))

	ret := map[string]interface{}{}
	if err := decoder.Decode(&ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func jwtDecode(s string) (map[string]interface{}, map[string]interface{}, error) {
	ss := strings.Split(s, ".")
	if len(ss) != 3 {
		return nil, nil, errors.New("invalid field number")
	}
	h, p, sig := ss[0], ss[1], ss[2]

	header, err := fieldDecode(h)
	if err != nil {
		return nil, nil, err
	}

	payload, err := fieldDecode(p)
	if err != nil {
		return nil, nil, err
	}

	var _ = sig

	return header, payload, nil
}

// Execute executes the command and returns an ExitStatus.
func (c *Command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	decoder := json.NewDecoder(os.Stdin)

	in := map[string]interface{}{}
	if err := decoder.Decode(&in); err != nil {
		log.Fatalf("Decode: %v", err)
	}

	res0, ok := in["AuthenticationResult"]
	if !ok {
		log.Fatalf("AuthenticationResult not found")
	}
	res, ok := res0.(map[string]interface{})
	if !ok {
		log.Fatalf("AuthenticationResult is invalid format")
	}

	hdr, p, err := jwtDecode(res["IdToken"].(string))
	if err != nil {
		log.Fatalf("DecodeError IdToken: %v", err)
	}
	for k, v := range hdr {
		log.Printf("IdToken header: %v = %v", k, v)
	}
	for k, v := range p {
		log.Printf("IdToken payload: %v = %v", k, v)
	}

	hdr, p, err = jwtDecode(res["AccessToken"].(string))
	if err != nil {
		log.Fatalf("DecodeError AccessToken: %v", err)
	}
	for k, v := range hdr {
		log.Printf("AccessToken header: %v = %v", k, v)
	}
	for k, v := range p {
		log.Printf("AccessToken payload: %v = %v", k, v)
	}

	log.Printf("RefreshToken: %v", res["RefreshToken"])

	return subcommands.ExitSuccess
}
