package kifudoc

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/golang/protobuf/jsonpb"

	"github.com/google/subcommands"

	"github.com/yunomu/kif"

	"github.com/yunomu/kansousen/lib/kifu"
	"github.com/yunomu/kansousen/lib/pbconv"
	documentpb "github.com/yunomu/kansousen/proto/document"
)

type Command struct {
	utf8 *bool

	marshaler *jsonpb.Marshaler
}

func NewCommand() *Command {
	return &Command{
		marshaler: &jsonpb.Marshaler{
			Indent:       "  ",
			EmitDefaults: true,
		},
	}
}

func (c *Command) Name() string     { return "kifudoc" }
func (c *Command) Synopsis() string { return "KIF(stdin) to document(stdout)" }
func (c *Command) Usage() string {
	return `
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	f.SetOutput(os.Stderr)

	c.utf8 = f.Bool("utf", false, "Input encoding UTF8")
}

func (c *Command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	in := os.Stdin

	var opts []kif.ParseOption
	if *c.utf8 {
		opts = append(opts, kif.ParseEncodingUTF8())
	}

	p := kifu.NewParser(kif.NewParser(opts...))
	kifu, steps, err := p.Parse(in, "test-user-id", "test-kifu-id")
	if err != nil {
		log.Fatalf("kifu.Parse: %v", err)
	}

	var docs []*documentpb.Document
	docs = append(docs, &documentpb.Document{
		Select: &documentpb.Document_Kifu{
			Kifu: kifu,
		},
	})
	for _, step := range steps {
		docs = append(docs, &documentpb.Document{
			Select: &documentpb.Document_Step{
				Step: step,
			},
		})
	}

	for _, pos := range pbconv.StepsToPositions(steps) {
		docs = append(docs, &documentpb.Document{
			Select: &documentpb.Document_Position{
				Position: pos,
			},
		})
	}

	for _, doc := range docs {
		if err := c.marshaler.Marshal(os.Stdout, doc); err != nil {
			log.Fatalf("jsonpb.Marshaler: %v", err)
		}
	}

	return subcommands.ExitSuccess
}
