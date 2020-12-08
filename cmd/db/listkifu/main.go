package listkifu

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/subcommands"

	"github.com/yunomu/kansousen/lib/db"
	documentpb "github.com/yunomu/kansousen/proto/document"
)

type Command struct {
	userId *string
}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Name() string     { return "listkifu" }
func (c *Command) Synopsis() string { return "List kifu" }
func (c *Command) Usage() string {
	return `
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	f.SetOutput(os.Stderr)

	c.userId = f.String("user-id", "", "User ID")
}

func (c *Command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	db := args[0].(func() db.DB)()

	if *c.userId == "" {
		log.Fatalf("user-id is required")
	}

	out := os.Stdout

	ctx, cancel := context.WithCancel(ctx)
	var rerr error
	if err := db.ListKifu(ctx, *c.userId, func(kifu *documentpb.Kifu) {
		if _, err := fmt.Fprintln(out, kifu.KifuId, kifu.Version); err != nil {
			rerr = err
			cancel()
		}
	}); err != nil {
		if rerr != nil {
			err = rerr
		}
		log.Fatalf("ListKifu: %v", err)
	}

	return subcommands.ExitSuccess
}
