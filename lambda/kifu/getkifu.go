package main

import (
	"context"

	"github.com/yunomu/kansousen/lib/lambda/lambdarpc"

	kifupb "github.com/yunomu/kansousen/proto/kifu"
)

func (h *handler) getKifu(ctx context.Context, reqCtx *lambdarpc.Context, in *kifupb.GetKifuRequest) (*kifupb.GetKifuResponse, error) {
	kifu, err := h.service.GetKifu(ctx, reqCtx.UserId, in.KifuId)
	if err != nil {
		return nil, err
	}

	var firstPlayers []*kifupb.GetKifuResponse_Player
	for _, p := range kifu.FirstPlayers {
		firstPlayers = append(firstPlayers, &kifupb.GetKifuResponse_Player{
			Name: p.Name,
			Note: p.Note,
		})
	}
	var secondPlayers []*kifupb.GetKifuResponse_Player
	for _, p := range kifu.SecondPlayers {
		secondPlayers = append(secondPlayers, &kifupb.GetKifuResponse_Player{
			Name: p.Name,
			Note: p.Note,
		})
	}

	return &kifupb.GetKifuResponse{
		UserId:        kifu.UserId,
		KifuId:        kifu.KifuId,
		StartTs:       kifu.Start.Unix(),
		Handicap:      kifu.Handicap,
		GameName:      kifu.GameName,
		FirstPlayers:  firstPlayers,
		SecondPlayers: secondPlayers,
		Note:          kifu.Note,
	}, nil
}
