package main

import (
	"context"

	"github.com/yunomu/kansousen/lib/lambda/lambdarpc"

	kifupb "github.com/yunomu/kansousen/proto/kifu"
)

func (h *handler) recentKifu(ctx context.Context, reqCtx *lambdarpc.Context, in *kifupb.RecentKifuRequest) (*kifupb.RecentKifuResponse, error) {
	kifus, err := h.service.RecentKifu(ctx, reqCtx.UserId, in.Limit)
	if err != nil {
		return nil, err
	}

	var ret []*kifupb.RecentKifuResponse_Kifu
	for _, k := range kifus {
		ret = append(ret, &kifupb.RecentKifuResponse_Kifu{
			UserId:        k.UserId,
			KifuId:        k.KifuId,
			StartTs:       k.Start.Unix(),
			Handicap:      k.Handicap,
			GameName:      k.GameName,
			FirstPlayers:  k.FirstPlayers,
			SecondPlayers: k.SecondPlayers,
			Note:          k.Note,
		})
	}

	return &kifupb.RecentKifuResponse{Kifus: ret}, nil
}
