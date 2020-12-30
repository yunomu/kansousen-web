package samepos

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
	pos    *string
	steps  *int
}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Name() string     { return "samepos" }
func (c *Command) Synopsis() string { return "Get same positions" }
func (c *Command) Usage() string {
	return `
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	f.SetOutput(os.Stderr)

	c.userId = f.String("user-id", "", "User ID")
	c.pos = f.String("pos", "", "SFEN position")
	c.steps = f.Int("steps", 5, "number of steps")
}

func (c *Command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if *c.userId == "" {
		log.Fatalf("user-id is required")
	}
	if *c.pos == "" {
		log.Fatalf("pos is required")
	}

	out := os.Stdout

	db := args[0].(func() db.DB)()

	ps, err := db.GetSamePositions(ctx, []string{*c.userId}, *c.pos, int32(*c.steps))
	if err != nil {
		log.Fatalf("GetSamePositions: %v", err)
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(ps); err != nil {
		log.Fatalf("json.Encode: %v", err)
	}

	return subcommands.ExitSuccess
}
