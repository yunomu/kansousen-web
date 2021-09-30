package db

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/google/subcommands"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/yunomu/kansousen/lib/db"

	"github.com/yunomu/kansousen/cmd/db/deletekifu"
	"github.com/yunomu/kansousen/cmd/db/getkifu"
	"github.com/yunomu/kansousen/cmd/db/listkifu"
	"github.com/yunomu/kansousen/cmd/db/putkifu"
	"github.com/yunomu/kansousen/cmd/db/recentkifu"
	"github.com/yunomu/kansousen/cmd/db/samepos"
)

type Command struct {
	endpoint *string
	region   *string
	table    *string
	log      *bool

	commander *subcommands.Commander
}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Name() string     { return "db" }
func (c *Command) Synopsis() string { return "db test" }
func (c *Command) Usage() string {
	return `
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	f.SetOutput(os.Stderr)

	c.endpoint = f.String("endpoint", "", "Endpoint of DynamoDB")
	c.region = f.String("region", "", "Endpoint of DynamoDB (default: config.json)")
	c.table = f.String("table", "", "Table name (default: config.json)")
	c.log = f.Bool("log", false, "output log")

	commander := subcommands.NewCommander(f, "")
	commander.Register(commander.FlagsCommand(), "help")
	commander.Register(commander.CommandsCommand(), "help")
	commander.Register(commander.HelpCommand(), "help")

	commander.Register(putkifu.NewCommand(), "kifu")
	commander.Register(getkifu.NewCommand(), "kifu")
	commander.Register(listkifu.NewCommand(), "kifu")
	commander.Register(deletekifu.NewCommand(), "kifu")
	commander.Register(recentkifu.NewCommand(), "kifu")

	commander.Register(samepos.NewCommand(), "pos")

	c.commander = commander
}

type logger struct{}

func (l *logger) Info(function string, message string) {
	log.Println(function, message)
}

func (c *Command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	cfg := args[0].(map[string]string)

	return c.commander.Execute(ctx, func() db.DB {
		region := cfg["Region"]
		if *c.region != "" {
			region = *c.region
		}
		config := aws.NewConfig().WithRegion(region)

		if *c.endpoint != "" {
			config.WithEndpoint(*c.endpoint)
		}

		return db.NewDynamoDB(
			dynamodb.New(session.New(), config),
			*c.table,
		)
	})
}
