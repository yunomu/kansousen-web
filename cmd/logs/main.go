package logs

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/subcommands"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type Command struct {
	profile      *string
	tz           *string
	region       *string
	logGroupName *string
}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) Name() string     { return "logs" }
func (c *Command) Synopsis() string { return "show latest lambda function logs" }
func (c *Command) Usage() string {
	return `
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	f.SetOutput(os.Stderr)

	c.profile = f.String("profile", "default", "AWS CLI profile")
	c.tz = f.String("tz", "Asia/Tokyo", "Time zone")
	c.region = f.String("region", "", "AWS region (default: config.json)")
	c.logGroupName = f.String("log-group-name", "", "Log group name for CloudWatch")
}

// Execute executes the command and returns an ExitStatus.
func (c *Command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	cfg := args[0].(map[string]string)

	loc, err := time.LoadLocation(*c.tz)
	if err != nil {
		log.Fatalf("LoadLocation: %v", err)
	}

	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: *c.profile,
	})
	if err != nil {
		log.Fatalf("NewSession: %v", err)
	}

	region := cfg["Region"]
	if *c.region != "" {
		region = *c.region
	}
	client := cloudwatchlogs.New(sess, aws.NewConfig().WithRegion(region))

	streams, err := client.DescribeLogStreamsWithContext(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(*c.logGroupName),
		OrderBy:      aws.String("LastEventTime"),
		Descending:   aws.Bool(true),
	})
	if err != nil {
		log.Fatalf("DescribeLogStreams: %v", err)
	}
	if len(streams.LogStreams) == 0 {
		log.Printf("empty")
		return subcommands.ExitSuccess
	}
	stream := streams.LogStreams[0]

	out, err := client.GetLogEventsWithContext(ctx, &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(*c.logGroupName),
		LogStreamName: stream.LogStreamName,
	})
	if err != nil {
		log.Fatalf("GetLogEvents: %v", err)
	}

	for _, event := range out.Events {
		ts := aws.Int64Value(event.Timestamp) / 1e3
		t := time.Unix(ts, 0).In(loc)
		fmt.Println(t.String())
		fmt.Println(aws.StringValue(event.Message))
	}

	return subcommands.ExitSuccess
}
