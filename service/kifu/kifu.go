package kifu

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"

	"github.com/yunomu/kif"

	"github.com/yunomu/kansousen/lib/db"
	libkifu "github.com/yunomu/kansousen/lib/kifu"
	"github.com/yunomu/kansousen/lib/pbconv"
	documentpb "github.com/yunomu/kansousen/proto/document"
)

type KifuServiceError interface {
	Error() string
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

type RecentKifu struct {
	UserId string
	KifuId string
	Start  time.Time

	Handicap      string
	GameName      string
	FirstPlayers  []string
	SecondPlayers []string
	Note          string
}

func (s *Service) RecentKifu(ctx context.Context, userId string, limit int32) ([]*RecentKifu, error) {
	kifus, err := s.table.GetRecentKifu(ctx, userId, int(limit))
	if err != nil {
		return nil, err
	}

	var ret []*RecentKifu
	for _, kifu := range kifus {
		tz := kifu.Timezone
		if tz == "" {
			tz = "Asia/Tokyo"
		}
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return nil, err
		}

		var firstPlayers, secondPlayers []string
		for _, player := range kifu.Players {
			switch player.Order {
			case documentpb.Player_BLACK:
				firstPlayers = append(firstPlayers, player.GetName())
			case documentpb.Player_WHITE:
				secondPlayers = append(secondPlayers, player.GetName())
			}
		}

		ret = append(ret, &RecentKifu{
			UserId: kifu.GetUserId(),
			KifuId: kifu.GetKifuId(),
			Start:  pbconv.DateTimeToTime(kifu.GetStart(), loc),

			Handicap:      kifu.GetHandicap().String(),
			GameName:      kifu.GetGameName(),
			FirstPlayers:  firstPlayers,
			SecondPlayers: secondPlayers,
			Note:          kifu.GetNote(),
		})
	}

	return ret, nil
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

type Format int

const (
	Format_UNKNOWN Format = iota
	Format_KIF
)

type Encoding int

const (
	Encoding_UTF8 Encoding = iota
	Encoding_SJIS
)

type DuplicatedKifu struct {
	UserId string
	KifuId string
}

func (k *DuplicatedKifu) String() string {
	return fmt.Sprintf("%s:%s", k.UserId, k.KifuId)
}

type DuplicatedKifuError struct {
	Kifus []*DuplicatedKifu
}

var _ KifuServiceError = (*DuplicatedKifuError)(nil)

func (e *DuplicatedKifuError) Error() string {
	var ds []string
	for _, k := range e.Kifus {
		ds = append(ds, k.String())
	}
	return fmt.Sprintf("duplicated kifu error: [%s]", strings.Join(ds, ","))
}

func (*DuplicatedKifuError) Type() string {
	return "DuplicatedKifuError"
}

type postKifuOptions struct {
	format                    Format
	encoding                  Encoding
	errorIfDuplicateSignature bool
}

type PostKifuOption func(*postKifuOptions)

func SetEncoding(encoding Encoding) PostKifuOption {
	return func(ops *postKifuOptions) {
		ops.encoding = encoding
	}
}

func SetFormat(format Format) PostKifuOption {
	return func(ops *postKifuOptions) {
		ops.format = format
	}
}

func SetErrorIfDuplicateSignature(flag bool) PostKifuOption {
	return func(ops *postKifuOptions) {
		ops.errorIfDuplicateSignature = flag
	}
}

func (s *Service) PostKifu(
	ctx context.Context,
	userId string,
	payload io.Reader,
	ops ...PostKifuOption,
) (string, error) {
	options := &postKifuOptions{
		format:   Format_KIF,
		encoding: Encoding_UTF8,
	}
	for _, f := range ops {
		f(options)
	}

	var parseOptions []kif.ParseOption
	switch options.encoding {
	case Encoding_UTF8:
		parseOptions = append(parseOptions, kif.ParseEncodingUTF8())
	case Encoding_SJIS:
		parseOptions = append(parseOptions, kif.ParseEncodingSJIS())
	default:
		return "", &InvalidArgumentError{
			typ: "UnknownEncodingError",
			msg: "unknown encoding error",
		}
	}

	switch options.format {
	case Format_KIF:
	default:
		return "", &InvalidArgumentError{
			typ: "UnknownFormatError",
			msg: "unknown format error",
		}
	}

	kifuUUID, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	parser := libkifu.NewParser(kif.NewParser(parseOptions...))

	kifu, steps, err := parser.Parse(payload, userId, kifuUUID.String())
	if err != nil {
		return "", err
	}

	if options.errorIfDuplicateSignature {
		sigs, err := s.table.DuplicateKifu(ctx, kifu.Sfen)
		if err != nil {
			return "", err
		}
		if len(sigs) > 0 {
			var dups []*DuplicatedKifu
			for _, sig := range sigs {
				dups = append(dups, &DuplicatedKifu{
					UserId: sig.UserId,
					KifuId: sig.KifuId,
				})
			}
			return "", &DuplicatedKifuError{
				Kifus: dups,
			}
		}
	}

	if err := s.table.PutKifu(ctx, kifu, steps); err != nil {
		return "", err
	}

	return kifuUUID.String(), nil
}

func (s *Service) DeleteKifu(ctx context.Context, userId string, kifuId string) error {
	return s.table.DeleteKifu(ctx, userId, kifuId)
}

type Pos struct {
	X int32
	Y int32
}

type Piece int32

const (
	Piece_NULL Piece = iota
	Piece_GYOKU
	Piece_HISHA
	Piece_RYU
	Piece_KAKU
	Piece_UMA
	Piece_KIN
	Piece_GIN
	Piece_NARI_GIN
	Piece_KEI
	Piece_NARI_KEI
	Piece_KYOU
	Piece_NARI_KYOU
	Piece_FU
	Piece_TO
)

type FinishedStatus int32

const (
	FinishedStatus_NOT_FINISHED FinishedStatus = iota
	FinishedStatus_SUSPEND
	FinishedStatus_SURRENDER
	FinishedStatus_DRAW
	FinishedStatus_REPETITION_DRAW
	FinishedStatus_CHECKMATE
	FinishedStatus_OVER_TIME_LIMIT
	FinishedStatus_FOUL_LOSS
	FinishedStatus_FOUL_WIN
	FinishedStatus_NYUGYOKU_WIN
)

type Step struct {
	Seq            int32
	Position       string
	Src            *Pos
	Dst            *Pos
	Piece          Piece
	FinishedStatus FinishedStatus
	Promoted       bool
	Captured       Piece
	TimestampSec   int32
	ThinkingSec    int32
	Notes          []string
}

type Player struct {
	Name string
	Note string
}

type Kifu struct {
	UserId string
	KifuId string

	Start         time.Time
	End           time.Time
	Handicap      string
	GameName      string
	FirstPlayers  []*Player
	SecondPlayers []*Player
	OtherFields   map[string]string
	Sfen          string
	CreatedTs     int64
	Steps         []*Step
	Note          string
}

func (s *Service) GetKifu(ctx context.Context, userId string, kifuId string) (*Kifu, error) {
	kifu, steps, err := s.table.GetKifuAndSteps(ctx, userId, kifuId)
	if err != nil {
		return nil, err
	}

	var resSteps []*Step
	for _, step := range steps {
		var resStep *Step

		resStep = &Step{
			Seq:          step.GetSeq(),
			Position:     step.GetPosition(),
			Promoted:     step.GetPromote(),
			Captured:     Piece(step.GetCaptured()),
			TimestampSec: step.GetTimestampSec(),
			ThinkingSec:  step.GetThinkingSec(),
			Notes:        step.Notes,

			FinishedStatus: FinishedStatus(step.GetFinishedStatus()),
		}

		if dst := step.GetDst(); dst != nil {
			resStep.Dst = &Pos{
				X: dst.GetX(),
				Y: dst.GetY(),
			}
		}
		resStep.Piece = Piece(step.GetPiece())
		if src := step.GetSrc(); src != nil {
			resStep.Src = &Pos{
				X: src.GetX(),
				Y: src.GetY(),
			}
		}

		resSteps = append(resSteps, resStep)
	}

	loc, err := time.LoadLocation(kifu.Timezone)
	if err != nil {
		return nil, err
	}

	var firstPlayers, secondPlayers []*Player
	for _, player := range kifu.Players {
		switch player.Order {
		case documentpb.Player_BLACK:
			firstPlayers = append(firstPlayers, &Player{
				Name: player.GetName(),
				Note: player.GetNote(),
			})
		case documentpb.Player_WHITE:
			secondPlayers = append(secondPlayers, &Player{
				Name: player.GetName(),
				Note: player.GetNote(),
			})
		}
	}

	var otherFields map[string]string
	for k, v := range kifu.OtherFields {
		otherFields[k] = v
	}

	return &Kifu{
		UserId: kifu.GetUserId(),
		KifuId: kifu.GetKifuId(),

		Start:         pbconv.DateTimeToTime(kifu.GetStart(), loc),
		End:           pbconv.DateTimeToTime(kifu.GetEnd(), loc),
		Handicap:      kifu.GetHandicap().String(),
		GameName:      kifu.GetGameName(),
		FirstPlayers:  firstPlayers,
		SecondPlayers: secondPlayers,
		OtherFields:   otherFields,
		Sfen:          kifu.GetSfen(),
		CreatedTs:     kifu.GetCreatedTs(),
		Steps:         resSteps,
		Note:          kifu.GetNote(),
	}, nil
}

type Phase_Step struct {
	Seq            int32
	Src            *Pos
	Dst            *Pos
	Piece          Piece
	Promoted       bool
	FinishedStatus FinishedStatus
}

type Phase struct {
	UserId string
	KifuId string
	Seq    int32
	Steps  []*Phase_Step
}

func (s *Service) GetSamePositions(
	ctx context.Context,
	userId string,
	position string,
	steps int32,
	excludeKifuIds []string,
) ([]*Phase, error) {
	pss, err := s.table.GetSamePositions(ctx,
		[]string{userId},
		position,
		db.GetSamePositionsSetNumStep(steps),
		db.GetSamePositionsAddExcludeKifuIds(excludeKifuIds),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "GetSamePositions: %v", err)
	}

	var kifus []*Phase
	for _, ps := range pss {
		for _, kid := range excludeKifuIds {
			if kid == ps.Position.GetKifuId() {
				continue
			}
		}

		var steps []*Phase_Step
		for _, step := range ps.Steps {
			var src, dst *Pos
			if step.GetDst() != nil {
				dst = &Pos{
					X: step.Dst.X,
					Y: step.Dst.Y,
				}
			}
			if step.GetSrc() != nil {
				src = &Pos{
					X: step.Src.X,
					Y: step.Src.Y,
				}
			}
			steps = append(steps, &Phase_Step{
				Seq:            step.GetSeq(),
				Dst:            dst,
				Src:            src,
				Piece:          Piece(step.GetPiece()),
				Promoted:       step.Promote,
				FinishedStatus: FinishedStatus(step.GetFinishedStatus()),
			})
		}

		kifus = append(kifus, &Phase{
			UserId: ps.Position.GetUserId(),
			KifuId: ps.Position.GetKifuId(),
			Seq:    ps.Position.GetSeq(),
			Steps:  steps,
		})
	}

	return kifus, nil
}
