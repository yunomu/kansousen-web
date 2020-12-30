package db

import (
	"context"
	"errors"

	documentpb "github.com/yunomu/kansousen/proto/document"
)

type getStepsOptions struct {
	start int32
	end   int32
}

type GetStepsOption func(*getStepsOptions)

func SetGetStepsRange(start, end int32) GetStepsOption {
	return func(o *getStepsOptions) {
		o.start = start
		o.end = end
	}
}

type PositionAndSteps struct {
	Position *documentpb.Position
	Steps    []*documentpb.Step
}

type DB interface {
	PutKifu(ctx context.Context, kifu *documentpb.Kifu, steps []*documentpb.Step) error
	GetKifu(ctx context.Context, userId, kifuId string) (*documentpb.Kifu, error)
	GetKifuAndSteps(ctx context.Context, userId, kifuId string) (*documentpb.Kifu, []*documentpb.Step, error)
	ListKifu(ctx context.Context, userId string, f func(*documentpb.Kifu)) error
	DuplicateKifu(ctx context.Context, sfen string) ([]*documentpb.KifuSignature, error)
	GetSteps(ctx context.Context, userId, kifuId string, options ...GetStepsOption) ([]*documentpb.Step, error)
	GetSamePositions(ctx context.Context, userIds []string, pos string, steps int32) ([]*PositionAndSteps, error)
	GetRecentKifu(ctx context.Context, userId string, limit int) ([]*documentpb.Kifu, error)
	DeleteKifu(ctx context.Context, userId, kifuId string) error
}

var (
	ErrEmpty        = errors.New("result is empty")
	ErrInvalidValue = errors.New("internal: invalid value")
)
