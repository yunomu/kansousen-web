package main

import (
	"context"

	"github.com/yunomu/kansousen/proto/lambdakifu"
)

func (h *handler) getKifu(ctx context.Context, in *lambdakifu.GetKifuInput) (*lambdakifu.GetKifuOutput, error) {
	kifu, err := h.service.GetKifu(ctx, in.UserId, in.KifuId)
	if err != nil {
		return nil, err
	}

	var firstPlayers []*lambdakifu.GetKifuOutput_Player
	for _, p := range kifu.FirstPlayers {
		firstPlayers = append(firstPlayers, &lambdakifu.GetKifuOutput_Player{
			Name: p.Name,
			Note: p.Note,
		})
	}
	var secondPlayers []*lambdakifu.GetKifuOutput_Player
	for _, p := range kifu.SecondPlayers {
		secondPlayers = append(secondPlayers, &lambdakifu.GetKifuOutput_Player{
			Name: p.Name,
			Note: p.Note,
		})
	}

	return &lambdakifu.GetKifuOutput{
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
