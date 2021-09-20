package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/yunomu/kif"

	"github.com/yunomu/kansousen/lib/db"
	libkifu "github.com/yunomu/kansousen/lib/kifu"
	"github.com/yunomu/kansousen/lib/lambda/lambdarpc"
	documentpb "github.com/yunomu/kansousen/proto/document"
	kifupb "github.com/yunomu/kansousen/proto/kifu"
)

type KifuServiceError interface {
	error
	Type() string
}

type Service struct {
	table db.DB
}

func NewService(table db.DB) *Service {
	return &Service{
		table: table,
	}
}

func (s *Service) RecentKifu(ctx context.Context, req *kifupb.RecentKifuRequest) (*kifupb.RecentKifuResponse, error) {
	userId := lambdarpc.GetUserId(ctx)

	kifus, err := s.table.GetRecentKifu(ctx, userId, int(req.GetLimit()))
	if err != nil {
		return nil, &lambdarpc.InternalError{
			Message: "db.GetRecentKifu",
			Err:     err,
		}
	}

	var ret []*kifupb.RecentKifuResponse_Kifu
	for _, kifu := range kifus {
		var firstPlayers, secondPlayers []string
		for _, player := range kifu.Players {
			switch player.Order {
			case documentpb.Player_BLACK:
				firstPlayers = append(firstPlayers, player.GetName())
			case documentpb.Player_WHITE:
				secondPlayers = append(secondPlayers, player.GetName())
			}
		}

		ret = append(ret, &kifupb.RecentKifuResponse_Kifu{
			UserId:  kifu.GetUserId(),
			KifuId:  kifu.GetKifuId(),
			StartTs: kifu.GetStartTs(),

			Handicap:      kifu.GetHandicap().String(),
			GameName:      kifu.GetGameName(),
			FirstPlayers:  firstPlayers,
			SecondPlayers: secondPlayers,
			Note:          kifu.GetNote(),
		})
	}

	return &kifupb.RecentKifuResponse{
		Kifus: ret,
	}, nil
}

type InvalidArgumentError struct {
	typ string
	msg string
}

var _ KifuServiceError = (*InvalidArgumentError)(nil)

func (e *InvalidArgumentError) Error() string {
	return e.msg
}

func (e *InvalidArgumentError) Type() string {
	return e.typ
}

func (s *Service) PostKifu(
	ctx context.Context,
	req *kifupb.PostKifuRequest,
) (*kifupb.PostKifuResponse, error) {
	userId := lambdarpc.GetUserId(ctx)
	if userId == "" {
		return nil, &lambdarpc.ClientError{
			Message: "UnauthorizedError",
		}
	}

	// XXX from request
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, &lambdarpc.ClientError{
			Message: "LoadLocation Asia/Tokyo",
			Err:     err,
		}
	}

	var parseOptions []kif.ParseOption
	switch req.Encoding {
	case "UTF-8":
		parseOptions = append(parseOptions, kif.ParseEncodingUTF8())
	case "Shift_JIS":
		parseOptions = append(parseOptions, kif.ParseEncodingSJIS())
	default:
		return nil, &lambdarpc.ClientError{
			Message: "UnknownEncodingError",
		}
	}

	switch req.Format {
	case "KIF":
	default:
		return nil, &lambdarpc.ClientError{
			Message: "UnknownFormatError",
		}
	}

	kifuUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, &lambdarpc.InternalError{
			Message: "uuid.NewRandom",
			Err:     err,
		}
	}

	parser := libkifu.NewParser(kif.NewParser(parseOptions...), loc)

	kifu, steps, err := parser.Parse(strings.NewReader(req.Payload), userId, kifuUUID.String())
	if err != nil {
		return nil, &lambdarpc.ClientError{
			Message: "kif parse error",
			Err:     err,
		}
	}

	version, err := s.table.PutKifu(ctx, kifu, steps, 0)
	if err != nil {
		return nil, &lambdarpc.InternalError{
			Message: "db.PutKifu",
			Err:     err,
		}
	}

	return &kifupb.PostKifuResponse{
		KifuId:  kifuUUID.String(),
		Version: version,
	}, nil
}

func (s *Service) DeleteKifu(ctx context.Context, req *kifupb.DeleteKifuRequest) (*kifupb.DeleteKifuResponse, error) {
	if err := s.table.DeleteKifu(ctx, req.GetKifuId(), req.GetVersion()); err != nil {
		return nil, &lambdarpc.InternalError{
			Message: "db.DeleteKifu",
			Err:     err,
		}
	}

	return &kifupb.DeleteKifuResponse{}, nil
}

func (s *Service) GetKifu(ctx context.Context, req *kifupb.GetKifuRequest) (*kifupb.GetKifuResponse, error) {
	kifu, steps, version, err := s.table.GetKifuAndSteps(ctx, req.GetKifuId())
	if err != nil {
		return nil, &lambdarpc.InternalError{
			Message: "db.GetKifuAndSteps",
			Err:     err,
		}
	}

	var resSteps []*kifupb.GetKifuResponse_Step
	for _, step := range steps {
		resStep := &kifupb.GetKifuResponse_Step{
			Seq:          step.GetSeq(),
			Position:     step.GetPosition(),
			Promoted:     step.GetPromote(),
			Captured:     kifupb.Piece_Id(step.GetCaptured()),
			TimestampSec: step.GetTimestampSec(),
			ThinkingSec:  step.GetThinkingSec(),
			Notes:        step.Notes,

			FinishedStatus: kifupb.FinishedStatus_Id(step.GetFinishedStatus()),
		}

		if dst := step.GetDst(); dst != nil {
			resStep.Dst = &kifupb.Pos{
				X: dst.GetX(),
				Y: dst.GetY(),
			}
		}
		resStep.Piece = kifupb.Piece_Id(step.GetPiece())
		if src := step.GetSrc(); src != nil {
			resStep.Src = &kifupb.Pos{
				X: src.GetX(),
				Y: src.GetY(),
			}
		}

		resSteps = append(resSteps, resStep)
	}

	var firstPlayers, secondPlayers []*kifupb.GetKifuResponse_Player
	for _, player := range kifu.Players {
		switch player.Order {
		case documentpb.Player_BLACK:
			firstPlayers = append(firstPlayers, &kifupb.GetKifuResponse_Player{
				Name: player.GetName(),
				Note: player.GetNote(),
			})
		case documentpb.Player_WHITE:
			secondPlayers = append(secondPlayers, &kifupb.GetKifuResponse_Player{
				Name: player.GetName(),
				Note: player.GetNote(),
			})
		}
	}

	var otherFields []*kifupb.Value
	for k, v := range kifu.OtherFields {
		otherFields = append(otherFields, &kifupb.Value{
			Name:  k,
			Value: v,
		})
	}

	return &kifupb.GetKifuResponse{
		UserId: kifu.GetUserId(),
		KifuId: kifu.GetKifuId(),

		StartTs:       kifu.GetStartTs(),
		EndTs:         kifu.GetEndTs(),
		Handicap:      kifu.GetHandicap().String(),
		GameName:      kifu.GetGameName(),
		FirstPlayers:  firstPlayers,
		SecondPlayers: secondPlayers,
		OtherFields:   otherFields,
		Sfen:          kifu.GetSfen(),
		CreatedTs:     kifu.GetCreatedTs(),
		Steps:         resSteps,
		Note:          kifu.GetNote(),
		Version:       version,
	}, nil
}

func (s *Service) GetSamePositions(ctx context.Context, req *kifupb.GetSamePositionsRequest) (*kifupb.GetSamePositionsResponse, error) {
	userId := lambdarpc.GetUserId(ctx)
	if userId == "" {
		return nil, &lambdarpc.ClientError{
			Message: "user-id is not found",
		}
	}

	pss, err := s.table.GetSamePositions(ctx,
		[]string{userId},
		req.GetPosition(),
		db.GetSamePositionsSetNumStep(req.GetSteps()),
		db.GetSamePositionsAddExcludeKifuIds(req.GetExcludeKifuIds()),
	)
	if err != nil {
		return nil, &lambdarpc.InternalError{
			Message: "GetSamePositions",
			Err:     err,
		}
	}

	var kifus []*kifupb.GetSamePositionsResponse_Kifu
	for _, ps := range pss {
		var steps []*kifupb.GetSamePositionsResponse_Step
		for _, step := range ps.Steps {
			var src, dst *kifupb.Pos
			if step.GetDst() != nil {
				dst = &kifupb.Pos{
					X: step.Dst.X,
					Y: step.Dst.Y,
				}
			}
			if step.GetSrc() != nil {
				src = &kifupb.Pos{
					X: step.Src.X,
					Y: step.Src.Y,
				}
			}
			steps = append(steps, &kifupb.GetSamePositionsResponse_Step{
				Seq:            step.GetSeq(),
				Dst:            dst,
				Src:            src,
				Piece:          kifupb.Piece_Id(step.GetPiece()),
				Promoted:       step.Promote,
				FinishedStatus: kifupb.FinishedStatus_Id(step.GetFinishedStatus()),
			})
		}

		kifus = append(kifus, &kifupb.GetSamePositionsResponse_Kifu{
			UserId: ps.UserId,
			KifuId: ps.KifuId,
			Steps:  steps,
		})
	}

	return &kifupb.GetSamePositionsResponse{
		Position: req.GetPosition(),
		Kifus:    kifus,
	}, nil
}
