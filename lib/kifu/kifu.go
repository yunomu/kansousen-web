package kifu

import (
	"io"
	"strings"

	"github.com/yunomu/kif"

	"github.com/yunomu/kansousen/lib/pbconv"
	documentpb "github.com/yunomu/kansousen/proto/document"
)

type Parser struct {
	kifParser *kif.Parser
}

func NewParser(kifParser *kif.Parser) *Parser {
	return &Parser{
		kifParser: kifParser,
	}
}

func (p *Parser) Parse(r io.Reader, userId, kifuId string) (*documentpb.Kifu, []*documentpb.Step, error) {
	k, err := p.kifParser.Parse(r)
	if err != nil {
		return nil, nil, err
	}

	kifu := &documentpb.Kifu{
		UserId: userId,
		KifuId: kifuId,
	}

	if err := pbconv.ReadHeader(k.Headers, kifu); err != nil {
		return nil, nil, err
	}

	var buf strings.Builder
	if err := kif.NewWriter(kif.SetFormat(kif.Format_SFEN)).Write(&buf, k); err != nil {
		return nil, nil, err
	}
	kifu.Sfen = buf.String()

	steps, err := pbconv.KifToSteps(kifu.UserId, kifu.KifuId, k)
	if err != nil {
		return nil, nil, err
	}

	return kifu, steps, nil
}
