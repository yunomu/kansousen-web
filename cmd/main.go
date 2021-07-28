package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/google/subcommands"

	"github.com/yunomu/kansousen/cmd/db"
	"github.com/yunomu/kansousen/cmd/decode"
	"github.com/yunomu/kansousen/cmd/kifudoc"
	"github.com/yunomu/kansousen/cmd/logs"
	"github.com/yunomu/kansousen/cmd/sfen"
)

var (
	utf8   = flag.Bool("utf", false, "Input encoding UTF8")
	config = flag.String("config", "static/config.json", "config.json")
)

func init() {
	log.SetOutput(os.Stderr)

	subcommands.Register(sfen.NewCommand(), "")
	subcommands.Register(decode.NewCommand(), "")
	subcommands.Register(kifudoc.NewCommand(), "")
	subcommands.Register(db.NewCommand(), "")
	subcommands.Register(logs.NewCommand(), "")

	subcommands.Register(subcommands.CommandsCommand(), "other")
	subcommands.Register(subcommands.FlagsCommand(), "other")
	subcommands.Register(subcommands.HelpCommand(), "other")

	flag.Parse()
}

func loadConfig(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	v := map[string]string{}
	d := json.NewDecoder(f)
	if err := d.Decode(&v); err != nil {
		return nil, err
	}

	return v, nil
}

func main() {
	ctx := context.Background()

	cfg, err := loadConfig(*config)
	if err != nil {
		log.Fatalln(err)
	}

	os.Exit(int(subcommands.Execute(ctx, cfg)))
}
