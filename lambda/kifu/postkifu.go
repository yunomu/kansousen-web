package main

import (
	"context"
	"strings"

	"github.com/yunomu/kansousen/service/kifu"

	"github.com/yunomu/kansousen/proto/lambdakifu"
)

type UnknownEncodingError struct {
	encoding lambdakifu.PostKifuInput_Encoding
}

func (e *UnknownEncodingError) Error() string {
	return "unknown encoding: " + e.encoding.String()
}

type UnknownFormatError struct {
	format lambdakifu.PostKifuInput_Format
}

func (e *UnknownFormatError) Error() string {
	return "unknown format: " + e.format.String()
}

func (h *handler) postKifu(ctx context.Context, in *lambdakifu.PostKifuInput) (*lambdakifu.PostKifuOutput, error) {
	var ops []kifu.PostKifuOption
	switch in.Encoding {
	case lambdakifu.PostKifuInput_UTF8:
		ops = append(ops, kifu.SetEncoding(kifu.Encoding_UTF8))
	case lambdakifu.PostKifuInput_SHIFT_JIS:
		ops = append(ops, kifu.SetEncoding(kifu.Encoding_SJIS))
	default:
		return nil, &UnknownEncodingError{encoding: in.Encoding}
	}

	switch in.Format {
	case lambdakifu.PostKifuInput_KIF:
		ops = append(ops, kifu.SetFormat(kifu.Format_KIF))
	default:
		return nil, &UnknownFormatError{format: in.Format}
	}

	kifuId, err := h.service.PostKifu(ctx, in.UserId, strings.NewReader(in.Payload), ops...)
	if err != nil {
		return nil, err
	}

	return &lambdakifu.PostKifuOutput{KifuId: kifuId}, nil
}
