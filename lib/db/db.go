package db

import (
	"context"
	"errors"

	documentpb "github.com/yunomu/kansousen/proto/document"
)

var (
	ErrUserIdIsEmpty   = errors.New("user_id is empty")
	ErrKifuIdIsEmpty   = errors.New("kifu_id is empty")
	ErrPositionIsEmpty = errors.New("position is empty")
)

type getStepsOptions struct {
	start int32
	end   int32
}

type GetStepsOption func(*getStepsOptions)

func GetStepsSetRange(start, end int32) GetStepsOption {
	return func(o *getStepsOptions) {
		o.start = start
		o.end = end
	}
}

type PositionAndSteps struct {
	Position *documentpb.Position
	Steps    []*documentpb.Step
}

type getSamePositionsOptions struct {
	numStep        int32
	excludeKifuIds []string
}

type GetSamePositionsOption func(*getSamePositionsOptions)

func GetSamePositionsSetNumStep(n int32) GetSamePositionsOption {
	return func(o *getSamePositionsOptions) {
		o.numStep = n
	}
}

func GetSamePositionsAddExcludeKifuId(kifuId string) GetSamePositionsOption {
	return func(o *getSamePositionsOptions) {
		o.excludeKifuIds = append(o.excludeKifuIds, kifuId)
	}
}

func GetSamePositionsAddExcludeKifuIds(kifuIds []string) GetSamePositionsOption {
	return func(o *getSamePositionsOptions) {
		o.excludeKifuIds = append(o.excludeKifuIds, kifuIds...)
	}
}

type DB interface {
	PutKifu(ctx context.Context, kifu *documentpb.Kifu, steps []*documentpb.Step) error
	GetKifu(ctx context.Context, userId, kifuId string) (*documentpb.Kifu, error)
	GetKifuAndSteps(ctx context.Context, userId, kifuId string) (*documentpb.Kifu, []*documentpb.Step, error)
	ListKifu(ctx context.Context, userId string, f func(*documentpb.Kifu)) error
	DuplicateKifu(ctx context.Context, sfen string) ([]*documentpb.KifuSignature, error)
	GetSteps(ctx context.Context, userId, kifuId string, options ...GetStepsOption) ([]*documentpb.Step, error)
	GetSamePositions(ctx context.Context, userIds []string, pos string, options ...GetSamePositionsOption) ([]*PositionAndSteps, error)
	GetRecentKifu(ctx context.Context, userId string, limit int) ([]*documentpb.Kifu, error)
	DeleteKifu(ctx context.Context, userId, kifuId string) error
}

var (
	ErrEmpty        = errors.New("result is empty")
	ErrInvalidValue = errors.New("internal: invalid value")
)
