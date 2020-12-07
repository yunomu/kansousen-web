package putkifu

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/google/subcommands"

	"github.com/yunomu/kif"

	"github.com/yunomu/kansousen/lib/db"
	"github.com/yunomu/kansousen/lib/kifu"
)

type Command struct {
	version *int64
	utf8    *bool
	userId  *string
	kifuId  *string
	dryrun  *bool
}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Name() string     { return "putkifu" }
func (c *Command) Synopsis() string { return "Put kif from stdin" }
func (c *Command) Usage() string {
	return `
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	f.SetOutput(os.Stderr)

	c.utf8 = f.Bool("utf", false, "Input encoding UTF8")
	c.userId = f.String("user-id", "", "User ID")
	c.kifuId = f.String("kifu-id", "", "Kifu ID")
	c.dryrun = f.Bool("dryrun", false, "Dry run")
}

func (c *Command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	db := args[0].(func() *db.DynamoDB)()

	if *c.userId == "" || *c.kifuId == "" {
		log.Fatalf("kifu-id and user-id is required")
	}

	in := os.Stdin

	var opts []kif.ParseOption
	if *c.utf8 {
		opts = append(opts, kif.ParseEncodingUTF8())
	}

	p := kifu.NewParser(kif.NewParser(opts...))

	kifu, steps, err := p.Parse(in, *c.userId, *c.kifuId)
	if err != nil {
		log.Fatalf("kifu.Parse: %v", err)
	}

	if !*c.dryrun {
		if err := db.PutKifu(ctx, kifu, steps); err != nil {
			log.Fatalf("PutKifu: %v", err)
		}
	}

	return subcommands.ExitSuccess
}
