package main

import (
	"context"
	"strings"

	"github.com/yunomu/kansousen/lib/lambda/lambdarpc"

	"github.com/yunomu/kansousen/service/kifu"

	kifupb "github.com/yunomu/kansousen/proto/kifu"
)

type UnknownEncodingError struct {
	encoding string
}

func (e *UnknownEncodingError) Error() string {
	return "unknown encoding: " + e.encoding
}

type UnknownFormatError struct {
	format string
}

func (e *UnknownFormatError) Error() string {
	return "unknown format: " + e.format
}

func (h *handler) postKifu(ctx context.Context, reqCtx *lambdarpc.Context, in *kifupb.PostKifuRequest) (*kifupb.PostKifuResponse, error) {
	var ops []kifu.PostKifuOption
	switch in.Encoding {
	case "UTF-8":
		ops = append(ops, kifu.SetEncoding(kifu.Encoding_UTF8))
	case "Shift_JIS":
		ops = append(ops, kifu.SetEncoding(kifu.Encoding_SJIS))
	default:
		return nil, &UnknownEncodingError{encoding: in.Encoding}
	}

	switch in.Format {
	case "KIF":
		ops = append(ops, kifu.SetFormat(kifu.Format_KIF))
	default:
		return nil, &UnknownFormatError{format: in.Format}
	}

	kifuId, err := h.service.PostKifu(ctx, reqCtx.UserId, strings.NewReader(in.Payload), ops...)
	if err != nil {
		return nil, err
	}

	return &kifupb.PostKifuResponse{KifuId: kifuId}, nil
}
