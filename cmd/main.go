package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/google/subcommands"

	"github.com/yunomu/kansousen/cmd/db"
	"github.com/yunomu/kansousen/cmd/decode"
	"github.com/yunomu/kansousen/cmd/kifudoc"
	"github.com/yunomu/kansousen/cmd/logs"
	"github.com/yunomu/kansousen/cmd/sfen"
	"github.com/yunomu/kansousen/cmd/signin"
)

var (
	utf8 = flag.Bool("utf", false, "Input encoding UTF8")
)

func init() {
	log.SetOutput(os.Stderr)

	subcommands.Register(sfen.NewCommand(), "")
	subcommands.Register(signin.NewCommand(), "")
	subcommands.Register(decode.NewCommand(), "")
	subcommands.Register(kifudoc.NewCommand(), "")
	subcommands.Register(db.NewCommand(), "")
	subcommands.Register(logs.NewCommand(), "")

	subcommands.Register(subcommands.CommandsCommand(), "other")
	subcommands.Register(subcommands.FlagsCommand(), "other")
	subcommands.Register(subcommands.HelpCommand(), "other")

	flag.Parse()
}

func main() {
	ctx := context.Background()

	os.Exit(int(subcommands.Execute(ctx)))
}
