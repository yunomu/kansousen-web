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
	version *int64
	utf8    *bool
	userId  *string
	kifuId  *string
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

	c.userId = f.String("user-id", "", "User ID")
	c.kifuId = f.String("kifu-id", "", "Kifu ID")
}

func (c *Command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	db := args[0].(func() *db.DynamoDB)()

	if *c.userId == "" || *c.kifuId == "" {
		log.Fatalf("kifu-id and user-id is required")
	}

	if err := db.DeleteKifu(ctx, *c.userId, *c.kifuId); err != nil {
		log.Fatalf("DeleteKifu: %v", err)
	}

	return subcommands.ExitSuccess
}
