package pbconv

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yunomu/kif"
	"github.com/yunomu/kif/ptypes"
	"github.com/yunomu/usi/sfen"

	"github.com/yunomu/kansousen/proto/document"
)

func KifFinishedStatusToStatus(st ptypes.FinishedStatus_Id) document.FinishedStatus_Id {
	// XXX id sync
	return document.FinishedStatus_Id(st)
}

func KifPieceToPiece(p ptypes.Piece_Id) document.Piece_Id {
	// XXX id sync
	return document.Piece_Id(p)
}

func KifPosToPos(p *ptypes.Pos) *document.Pos {
	if p == nil || p.X == 0 || p.Y == 0 {
		return nil
	}

	return &document.Pos{
		X: p.X,
		Y: p.Y,
	}
}

func posXFromInt(x int32) sfen.PosX {
	x = 9 - x
	if x < 0 {
		panic("invalid x")
	}
	return sfen.PosXs[x]
}

func posYFromInt(y int32) sfen.PosY {
	if y <= 0 {
		panic("invalid y")
	}
	y -= 1
	return sfen.PosYs[y]
}

func stepToMove(p *sfen.Surface, step *ptypes.Step) *document.Move {
	var captured document.Piece_Id
	if step.FinishedStatus == ptypes.FinishedStatus_NOT_FINISHED {
		if piece := p.GetPiece(posXFromInt(step.Dst.X), posYFromInt(step.Dst.Y)); piece != nil {
			switch piece.Type {
			case sfen.Piece_NULL:
				captured = document.Piece_NULL
			case sfen.Piece_HISHA:
				captured = document.Piece_HISHA
			case sfen.Piece_KAKU:
				captured = document.Piece_KAKU
			case sfen.Piece_KIN:
				captured = document.Piece_KIN
			case sfen.Piece_GIN:
				captured = document.Piece_GIN
			case sfen.Piece_KEI:
				captured = document.Piece_KEI
			case sfen.Piece_KYOU:
				captured = document.Piece_KYOU
			case sfen.Piece_FU:
				captured = document.Piece_FU
			default:
				panic("Unknown Piece type")
			}
		}
	}

	return &document.Move{
		Src:            KifPosToPos(step.GetSrc()),
		Dst:            KifPosToPos(step.GetDst()),
		Piece:          KifPieceToPiece(step.GetPiece()),
		Promote:        step.GetModifier() == ptypes.Modifier_PROMOTE,
		Drop:           step.GetModifier() == ptypes.Modifier_PUTTED,
		Captured:       captured,
		FinishedStatus: KifFinishedStatusToStatus(step.GetFinishedStatus()),
	}
}

func KifToSteps(userId, kifuId string, k *ptypes.Kif) ([]*document.Step, error) {
	p := sfen.NewSurfaceStartpos()
	var steps []*document.Step

	var buf strings.Builder
	if err := p.PrintSFEN(&buf); err != nil {
		return nil, err
	}
	steps = append(steps, &document.Step{
		UserId: userId,
		KifuId: kifuId,
		Seq:    0,

		Position:     buf.String(),
		Move:         nil,
		TimestampSec: 0,
		ThinkingSec:  0,
		Notes:        nil,
	})

	for _, step := range k.Steps {
		s := &document.Step{
			UserId: userId,
			KifuId: kifuId,
			Seq:    step.GetSeq(),

			TimestampSec: step.GetElapsedSec(),
			ThinkingSec:  step.GetThinkingSec(),
			Notes:        step.GetNotes(),
		}

		move := kif.StepToMove(step)
		s.Move = stepToMove(p, step)
		s.Move.Sfen = move

		if move != "" {
			if err := p.Move(move); err != nil {
				return nil, err
			}

			buf.Reset()
			p.SetStep(1)
			if err := p.PrintSFEN(&buf); err != nil {
				return nil, err
			}
		}
		s.Position = buf.String()

		steps = append(steps, s)

		if step.FinishedStatus != ptypes.FinishedStatus_NOT_FINISHED {
			break
		}
	}

	return steps, nil
}

func StepsToPositions(steps []*document.Step) []*document.Position {
	var moves []*document.Move
	for _, step := range steps {
		if step.Seq == 0 {
			// and pos = startpos
			continue
		}
		if step.Move == nil {
			break
		}
		moves = append(moves, step.Move)
	}

	var positions []*document.Position
	for i, step := range steps {
		if step.Seq == 0 {
			continue
		}
		var ms []*document.Move
		if i < len(moves) {
			ms = moves[i:]
		}
		positions = append(positions, &document.Position{
			UserId:   step.UserId,
			Position: step.Position,
			KifuId:   step.KifuId,
			Seq:      step.Seq,

			Moves: ms,
		})
	}

	return positions
}

var handicapString = []string{
	"平手",
	"香落ち",
	"右香落ち",
	"角落ち",
	"飛車落ち",
	"飛香落ち",
	"二枚落ち",
	"三枚落ち",
	"四枚落ち",
	"五枚落ち",
	"左五枚落ち",
	"六枚落ち",
	"八枚落ち",
	"十枚落ち",
	"その他",
}

func HandicapString(h document.Handicap_Id) string {
	if i := int(h); i >= 0 && i < len(handicapString) {
		return handicapString[i]
	}

	panic("unknown handicap")
}

func ParseHandicap(s string) document.Handicap_Id {
	if s == "" {
		return document.Handicap_NONE
	}

	for i, str := range handicapString {
		if str == s {
			return document.Handicap_Id(i)
		}
	}
	return document.Handicap_OTHER
}

func ParseDateTime(s string) (*document.DateTime, error) {
	r, err := regexp.Compile(
		`(\d{4})(?:[/年])(\d{2})(?:[/月])(\d{2})日?(?:\([日月火水木金土]\))?( (\d{2})[:：](\d{2})[:：](\d{2}))?`,
	)
	if err != nil {
		return nil, err
	}

	ss := r.FindStringSubmatch(s)
	l := len(ss)
	if l < 4 {
		return nil, fmt.Errorf("parse error: field number is mismatch: len=%v", l)
	}

	year, err := strconv.ParseInt(ss[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse error: year")
	}
	month, err := strconv.ParseInt(ss[2], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse error: month")
	}
	day, err := strconv.ParseInt(ss[3], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse error: day")
	}

	ret := &document.DateTime{
		Date: &document.Date{
			Year:  int32(year),
			Month: int32(month),
			Day:   int32(day),
		},
	}

	if len(ss[4]) == 0 {
		return ret, nil
	}

	hour, err := strconv.ParseInt(ss[5], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse error: hour")
	}
	min, err := strconv.ParseInt(ss[6], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse error: minute")
	}
	sec, err := strconv.ParseInt(ss[7], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse error: second")
	}

	ret.TimeSec = int32(hour*60*60 + min*60 + sec)

	return ret, nil
}

func ReadHeader(hs []*ptypes.Header, out *document.Kifu) error {
	header := map[string]string{}
	used := map[string]struct{}{}
	for _, h := range hs {
		header[h.Name] = h.Value
	}

	fieldfs := []struct {
		field string
		f     func(field, v string) error
	}{
		{
			field: "開始日時",
			f: func(field, v string) error {
				if v == "" {
					return nil
				}

				start, err := ParseDateTime(v)
				if err != nil {
					return err
				}

				out.Start = start
				return nil
			},
		},
		{
			field: "対局日",
			f: func(field, v string) error {
				if v == "" {
					return nil
				}

				start, err := ParseDateTime(v)
				if err != nil {
					return err
				}

				out.Start = start
				return nil
			},
		},
		{
			field: "終了日時",
			f: func(field, v string) error {
				if v == "" {
					return nil
				}

				start, err := ParseDateTime(v)
				if err != nil {
					return err
				}

				out.End = start
				return nil
			},
		},
		{
			field: "",
			f: func(field, v string) error {
				out.Handicap = ParseHandicap(header["手割合"])
				used["手割合"] = struct{}{}
				return nil
			},
		},
		{
			field: "棋戦",
			f: func(field, v string) error {
				out.GameName = v
				used[field] = struct{}{}
				return nil
			},
		},
		{
			field: "先手",
			f: func(field, v string) error {
				out.Players = append(out.Players, &document.Player{
					Order: document.Player_BLACK,
					Name:  v,
				})
				return nil
			},
		},
		{
			field: "上手",
			f: func(field, v string) error {
				out.Players = append(out.Players, &document.Player{
					Order: document.Player_BLACK,
					Name:  v,
				})
				return nil
			},
		},
		{
			field: "後手",
			f: func(field, v string) error {
				out.Players = append(out.Players, &document.Player{
					Order: document.Player_WHITE,
					Name:  v,
				})
				return nil
			},
		},
		{
			field: "下手",
			f: func(field, v string) error {
				out.Players = append(out.Players, &document.Player{
					Order: document.Player_WHITE,
					Name:  v,
				})
				return nil
			},
		},
	}

	for _, fieldf := range fieldfs {
		var val string
		if fieldf.field != "" {
			v, ok := header[fieldf.field]
			if !ok {
				continue
			}
			val = v
		}

		if err := fieldf.f(fieldf.field, val); err != nil {
			return err
		}

		used[fieldf.field] = struct{}{}
	}

	out.OtherFields = make(map[string]string)
	for k, v := range header {
		if _, ok := header[k]; ok {
			continue
		}

		out.OtherFields[k] = v
	}

	return nil
}

func DateTimeToTS(dt *document.DateTime, loc *time.Location) int64 {
	d := dt.GetDate()
	tsec := int(dt.GetTimeSec())
	sec := tsec % 60
	tmin := tsec / 60
	min := tmin % 60
	hour := tmin / 60
	t := time.Date(int(d.GetYear()), time.Month(d.GetMonth()), int(d.GetDay()), hour, min, sec, 0, loc)
	return t.Unix()
}
