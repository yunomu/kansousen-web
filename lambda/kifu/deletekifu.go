package main

import (
	"context"

	"github.com/yunomu/kansousen/proto/lambdakifu"
)

func (h *handler) deleteKifu(ctx context.Context, in *lambdakifu.DeleteKifuInput) error {
	return h.service.DeleteKifu(ctx, in.UserId, in.KifuId)
}
