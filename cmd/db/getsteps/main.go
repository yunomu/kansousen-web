package getsteps

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
	limit  *int
	start  *int
}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Name() string     { return "getsteps" }
func (c *Command) Synopsis() string { return "Get steps" }
func (c *Command) Usage() string {
	return `
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	f.SetOutput(os.Stderr)

	c.userId = f.String("user-id", "", "User ID")
	c.kifuId = f.String("kifu-id", "", "Kifu ID")
	c.limit = f.Int("limit", 0, "Limit")
	c.start = f.Int("start", 0, "start with")
}

func (c *Command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if *c.userId == "" {
		log.Fatalf("user-id is required")
	}
	if *c.kifuId == "" {
		log.Fatalf("kifu-id is required")
	}

	out := os.Stdout

	var opts []db.GetStepsOption
	if *c.start > 0 {
		opts = append(opts, db.SetGetStepsRange(int32(*c.start), int32(*c.start+*c.limit)))
	}

	db := args[0].(func() db.DB)()
	steps, err := db.GetSteps(ctx, *c.userId, *c.kifuId, opts...)
	if err != nil {
		log.Fatalf("GetSteps: %v", err)
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(steps); err != nil {
		log.Fatalf("json.Encode: %v", err)
	}

	return subcommands.ExitSuccess
}
