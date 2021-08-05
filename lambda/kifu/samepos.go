package main

import (
	"context"

	"github.com/yunomu/kansousen/lib/lambda/requestcontext"

	kifupb "github.com/yunomu/kansousen/proto/kifu"
)

func (h *handler) getSamePositions(ctx context.Context, reqCtx *requestcontext.Context, in *kifupb.GetSamePositionsRequest) (*kifupb.GetSamePositionsResponse, error) {
	phases, err := h.service.GetSamePositions(ctx, reqCtx.UserId, in.Position, in.Steps, in.ExcludeKifuIds)
	if err != nil {
		return nil, err
	}

	var ret []*kifupb.GetSamePositionsResponse_Kifu
	for _, p := range phases {
		var steps []*kifupb.GetSamePositionsResponse_Step
		for _, s := range p.Steps {
			steps = append(steps, &kifupb.GetSamePositionsResponse_Step{
				Seq: s.Seq,
				Src: &kifupb.Pos{
					X: s.Src.X,
					Y: s.Src.Y,
				},
				Dst: &kifupb.Pos{
					X: s.Dst.X,
					Y: s.Dst.Y,
				},
				Piece:          kifupb.Piece_Id(s.Piece),
				Promoted:       s.Promoted,
				FinishedStatus: kifupb.FinishedStatus_Id(s.FinishedStatus),
			})
		}
		ret = append(ret, &kifupb.GetSamePositionsResponse_Kifu{
			UserId: p.UserId,
			KifuId: p.KifuId,
			Seq:    p.Seq,
			Steps:  steps,
		})
	}

	return &kifupb.GetSamePositionsResponse{
		Position: in.Position,
		Kifus:    ret,
	}, nil
}
