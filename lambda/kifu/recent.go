package main

import (
	"context"

	"github.com/yunomu/kansousen/proto/lambdakifu"
)

func (h *handler) recentKifu(ctx context.Context, in *lambdakifu.RecentKifuInput) (*lambdakifu.RecentKifuOutput, error) {
	kifus, err := h.service.RecentKifu(ctx, in.UserId, in.Limit)
	if err != nil {
		return nil, err
	}

	var ret []*lambdakifu.RecentKifuOutput_Kifu
	for _, k := range kifus {
		ret = append(ret, &lambdakifu.RecentKifuOutput_Kifu{
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

	return &lambdakifu.RecentKifuOutput{Kifus: ret}, nil
}
