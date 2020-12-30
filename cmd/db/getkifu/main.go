package getkifu

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/google/subcommands"

	"github.com/yunomu/kansousen/lib/db"
)

type Command struct {
	userId *string
	kifuId *string
}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Name() string     { return "getkifu" }
func (c *Command) Synopsis() string { return "Get kifu" }
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
	if *c.userId == "" {
		log.Fatalf("user-id is required")
	}
	if *c.kifuId == "" {
		log.Fatalf("kifu-id is required")
	}

	db := args[0].(func() db.DB)()

	out := os.Stdout

	kifu, err := db.GetKifu(ctx, *c.userId, *c.kifuId)
	if err != nil {
		log.Fatalf("ListKifu: %v", err)
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(kifu); err != nil {
		log.Fatalf("ListKifu: %v", err)
	}

	return subcommands.ExitSuccess
}
