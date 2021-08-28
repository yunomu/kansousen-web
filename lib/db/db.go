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

type Position struct {
	UserId string
	KifuId string
	Steps  []*documentpb.Step
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

type UserKifu struct {
	UserId string
	KifuId string
}

type DB interface {
	PutKifu(ctx context.Context, kifu *documentpb.Kifu, steps []*documentpb.Step, version int64) error
	GetKifu(ctx context.Context, kifuId string) (*documentpb.Kifu, int64, error)
	GetKifuAndSteps(ctx context.Context, kifuId string) (*documentpb.Kifu, []*documentpb.Step, int64, error)
	ListKifu(ctx context.Context, userId string, f func(*documentpb.Kifu, int64)) error
	GetKifuIdsBySfen(ctx context.Context, sfen string) ([]*UserKifu, error)
	GetSamePositions(ctx context.Context, userIds []string, pos string, options ...GetSamePositionsOption) ([]*Position, error)
	GetRecentKifu(ctx context.Context, userId string, limit int) ([]*documentpb.Kifu, error)
	DeleteKifu(ctx context.Context, userId, kifuId string) error
}

var (
	ErrEmpty = errors.New("result is empty")
)

type ErrInvalidValue struct {
	Details string
}

func (*ErrInvalidValue) Error() string {
	return "invalid value"
}
