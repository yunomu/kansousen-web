package kifudoc

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/google/subcommands"

	"github.com/yunomu/kif"

	"github.com/yunomu/kansousen/lib/kifu"
)

type Command struct {
	utf8 *bool
	tz   *string
}

func NewCommand() *Command {
	return &Command{}
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
	c.tz = f.String("tz", "Asia/Tokyo", "Timezone")
}

func (c *Command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	loc, err := time.LoadLocation(*c.tz)
	if err != nil {
		log.Fatalf("LoadLocation: %v", err)
	}

	in := os.Stdin

	var opts []kif.ParseOption
	if *c.utf8 {
		opts = append(opts, kif.ParseEncodingUTF8())
	}

	p := kifu.NewParser(kif.NewParser(opts...), loc)
	kifu, steps, err := p.Parse(in, "test-user-id", "test-kifu-id")
	if err != nil {
		log.Fatalf("kifu.Parse: %v", err)
	}

	out := map[string]interface{}{}
	out["kifu"] = kifu
	out["steps"] = steps

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		log.Fatalf("json.Encode: %v", err)
	}

	return subcommands.ExitSuccess
}
