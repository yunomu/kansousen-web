package recentkifu

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/subcommands"

	"github.com/yunomu/kansousen/lib/db"
)

type Command struct {
	userId *string
}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Name() string     { return "recentkifu" }
func (c *Command) Synopsis() string { return "get recent kif from stdin" }
func (c *Command) Usage() string {
	return `
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	f.SetOutput(os.Stderr)

	c.userId = f.String("user-id", "", "User ID")
}

func (c *Command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	db := args[0].(func() *db.DynamoDB)()

	if *c.userId == "" {
		log.Fatalf("user-id is required")
	}

	kifus, err := db.GetRecentKifu(ctx, *c.userId)
	if err != nil {
		log.Fatalf("GetRecentKifu: %v", err)
	}

	for _, kifu := range kifus {
		fmt.Printf("%s", kifu.KifuId)
		fmt.Println()
	}
	fmt.Println("len:", len(kifus))

	return subcommands.ExitSuccess
}
