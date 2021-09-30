package deletekifu

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/google/subcommands"

	"github.com/yunomu/kansousen/lib/db"
)

type Command struct {
	utf8    *bool
	kifuId  *string
	version *int64
	dryrun  *bool
}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Name() string     { return "deletekifu" }
func (c *Command) Synopsis() string { return "Delete kif from stdin" }
func (c *Command) Usage() string {
	return `
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	f.SetOutput(os.Stderr)

	c.kifuId = f.String("kifu-id", "", "Kifu ID")
	c.version = f.Int64("version", 0, "Version")
}

func (c *Command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	db := args[0].(func() db.DB)()

	if *c.version == 0 || *c.kifuId == "" {
		log.Fatalf("kifu-id and version is required")
	}

	if err := db.DeleteKifu(ctx, *c.kifuId, *c.version); err != nil {
		log.Fatalf("DeleteKifu: %v", err)
	}

	return subcommands.ExitSuccess
}
