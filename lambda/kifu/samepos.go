package main

import (
	"context"

	"github.com/yunomu/kansousen/proto/lambdakifu"
)

func (h *handler) getSamePositions(ctx context.Context, in *lambdakifu.GetSamePositionsInput) (*lambdakifu.GetSamePositionsOutput, error) {
	phases, err := h.service.GetSamePositions(ctx, in.UserId, in.Position, in.Steps, in.ExcludeKifuIds)
	if err != nil {
		return nil, err
	}

	var ret []*lambdakifu.GetSamePositionsOutput_Phase
	for _, p := range phases {
		var steps []*lambdakifu.GetSamePositionsOutput_Step
		for _, s := range p.Steps {
			steps = append(steps, &lambdakifu.GetSamePositionsOutput_Step{
				Seq: s.Seq,
				Src: &lambdakifu.Pos{
					X: s.Src.X,
					Y: s.Src.Y,
				},
				Dst: &lambdakifu.Pos{
					X: s.Dst.X,
					Y: s.Dst.Y,
				},
				Piece:          lambdakifu.Piece_Id(s.Piece),
				Promoted:       s.Promoted,
				FinishedStatus: lambdakifu.FinishedStatus_Id(s.FinishedStatus),
			})
		}
		ret = append(ret, &lambdakifu.GetSamePositionsOutput_Phase{
			UserId: p.UserId,
			KifuId: p.KifuId,
			Seq:    p.Seq,
			Steps:  steps,
		})
	}

	return &lambdakifu.GetSamePositionsOutput{
		Position: in.Position,
		Phases:   ret,
	}, nil
}
