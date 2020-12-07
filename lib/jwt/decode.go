package jwt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

var (
	ErrInvalidNumberOfSections = errors.New("invalid number of sections")
)

func decodeSection(s string) (map[string]interface{}, error) {
	bs, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	d := json.NewDecoder(bytes.NewReader(bs))

	ret := map[string]interface{}{}
	if err := d.Decode(&ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func DecodePayload(s string) (map[string]interface{}, error) {
	ss := strings.Split(s, ".")
	if len(ss) != 3 {
		return nil, ErrInvalidNumberOfSections
	}

	return decodeSection(ss[1])
}
