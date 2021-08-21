package main

import (
	"context"

	"github.com/yunomu/kansousen/lib/lambda/lambdarpc"

	"github.com/yunomu/kansousen/proto/kifu"
)

func (h *handler) deleteKifu(ctx context.Context, reqCtx *lambdarpc.Context, in *kifu.DeleteKifuRequest) error {
	return h.service.DeleteKifu(ctx, reqCtx.UserId, in.KifuId)
}
