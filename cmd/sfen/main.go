package sfen

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/google/subcommands"

	"github.com/yunomu/kif"

	"github.com/yunomu/kansousen/lib/pbconv"
)

type Command struct {
	utf8 *bool
}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Name() string     { return "sfen" }
func (c *Command) Synopsis() string { return "KIF(stdin) to SFEN(stdout)" }
func (c *Command) Usage() string {
	return `
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	f.SetOutput(os.Stderr)

	c.utf8 = f.Bool("utf", false, "Input encoding UTF8")
}

// Execute executes the command and returns an ExitStatus.
func (c *Command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	in := os.Stdin

	var opts []kif.ParseOption
	if *c.utf8 {
		opts = append(opts, kif.ParseEncodingUTF8())
	}

	p := kif.NewParser(opts...)
	k, err := p.Parse(in)
	if err != nil {
		log.Fatalf("kif.Parse: %v", err)
	}

	w := kif.NewWriter(kif.SetFormat(kif.Format_SFEN))
	var buf strings.Builder
	if err := w.Write(&buf, k); err != nil {
		log.Fatalf("kif.Write: %v", err)
	}

	steps, err := pbconv.KifToSteps("", "", k)
	if err != nil {
		log.Fatalf("pbconv.KifToSteps: %v", err)
	}

	log.Println(steps)

	return subcommands.ExitSuccess
}
